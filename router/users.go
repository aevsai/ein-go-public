package router

import (
	"bytes"
	"context"
	"einstein-server/database"
	"einstein-server/oauth"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/awa/go-iap/appstore/api"
	"github.com/google/uuid"
)

type Transaction struct {
    BundleID                  string  `json:"bundleId"`
    DeviceVerification        string  `json:"deviceVerification"`
    DeviceVerificationNonce   string  `json:"deviceVerificationNonce"`
    Environment               string  `json:"environment"`
    ExpiresDate               float64 `json:"expiresDate"`
    InAppOwnershipType        string  `json:"inAppOwnershipType"`
    IsUpgraded                bool    `json:"isUpgraded"`
    OriginalPurchaseDate      float64 `json:"originalPurchaseDate"`
    OriginalTransactionID     string  `json:"originalTransactionId"`
    ProductID                 string  `json:"productId"`
    PurchaseDate              float64 `json:"purchaseDate"`
    Quantity                  int     `json:"quantity"`
    SignedDate                float64 `json:"signedDate"`
    SubscriptionGroupID       string  `json:"subscriptionGroupIdentifier"`
    TransactionID             string  `json:"transactionId"`
    Type                      string  `json:"type"`
    WebOrderLineItemID        string  `json:"webOrderLineItemId"`
}

func SafeMatch(pattern string, path string) bool {
    matched, err := regexp.MatchString(pattern, path)

    if err != nil {
        logger.Printf("Error matching path: %s \n; %s", path, err)
        return false
    }
    return matched
}

func HandleLinkUser(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        buf := bytes.NewBuffer(nil)
        if _, err := io.Copy(buf, r.Body); err == nil {
            token := string(buf.Bytes())
            if oauth.ValidateToken(token) {
                user, _ := GetUser(r)
                client, err := oauth.GetAuthClient()
                if  err != nil {
                    logger.Err(err).Msg("")
                    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
                    return
                }
                userInfo, err := client.UserInfo(context.TODO(), token)
                if  err != nil {
                    logger.Err(err).Msg("")
                    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
                    return
                }
                db := database.GetConnection()
                defer db.Close()
                var secondaryUser database.User 
                logger.Debug().Msgf("%+v", userInfo)
                if err = db.Get(&secondaryUser, database.SqlUserSelectByExtId, userInfo.Sub); err != nil {
                    logger.Err(err).Msg("")
                    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
                    return
                }
                secondaryUser.ParentID = user.ID
                if _, err := db.NamedExec(database.SqlUserUpdate, secondaryUser); err != nil {
                    logger.Err(err).Msg("")
                    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
                    return
                }
            } else {
                logger.Err(err).Msg("")
                http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
                return
            }
        }  else {
            logger.Err(err).Msg("")
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
            return
        }

    } else {
        http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
        return
    }
}

func HandleVerify(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
        //user, _ := GetUser(r)
        var inTransaction Transaction
        if err := json.NewDecoder(r.Body).Decode(&inTransaction); err != nil {
            logger.Err(err).Msg("")
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        file, err := os.Open("/app/cert.p8")
        if err != nil {
            logger.Err(err).Msg("")
            http.Error(w, "Internal server error.", http.StatusInternalServerError)
            return
        }
        buf := bytes.NewBuffer(nil)
        if _, err = buf.ReadFrom(file); err != nil {
            logger.Err(err).Msg("")
            http.Error(w, "Internal server error.", http.StatusInternalServerError)
            return
        }
        var isSandbox = false
        if inTransaction.Environment == "Sandbox" {
            isSandbox = true
        } 
        c := &api.StoreConfig{
            KeyContent: buf.Bytes(),  // Loads a .p8 certificate
            KeyID:      "6N65L53RN6",                // Your private key ID from App Store Connect (Ex: 2X9R4HXF34)
            BundleID:   inTransaction.BundleID,           // Your app’s bundle ID
            Issuer:     "d221f7af-feac-4cc2-96c7-e437cfe3d44a",  // Your issuer ID from the Keys page in App Store Connect (Ex: "57246542-96fe-1a63-e053-0824d011072a")
            Sandbox: isSandbox,
        }
        a := api.NewStoreClient(c)
        ctx := context.Background()
        response, err := a.GetTransactionInfo(ctx, inTransaction.TransactionID)
        if err != nil {
             // error handling
            logger.Err(err).Msg("")
            http.Error(w, "Internal server error.", http.StatusInternalServerError)
            return
        }
        transaction, err := a.ParseSignedTransaction(response.SignedTransactionInfo)
        if err != nil {
            // error handling
            logger.Err(err).Msg("")
            http.Error(w, "Internal server error.", http.StatusInternalServerError)
            return
        }
        if transaction.TransactionID == inTransaction.TransactionID {
            if transaction.InAppOwnershipType == "PURCHASED" {
                user, _ := GetUser(r)
                user.Subscribed = true
                db := database.GetConnection()
                defer db.Close()
                _, err := db.NamedExec(database.SqlUserUpdate, user)
                if err != nil {
                    logger.Err(err)
                    http.Error(w, "Internal server error.", http.StatusInternalServerError)
                    return
                }
                db.Close()
                json.NewEncoder(w).Encode(user)
            }
        }
    }
}

func HandleUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case SafeMatch("/users(/)*", r.URL.Path):
		switch r.Method {
		case http.MethodGet:
			user, _ := GetUser(r)
            logger.Printf("%+v", user)
        	json.NewEncoder(w).Encode(user)
		case http.MethodPost:
			var newUser database.User
			if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
				logger.Err(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			currUser, extId := GetUser(r)
			if currUser.ID.UUID != uuid.Nil {
				// Update existing user
				db := database.GetConnection()
				currUser.Email = newUser.Email
				currUser.Name = newUser.Name
				currUser.ProfilePicture = newUser.ProfilePicture
				db.NamedExec(database.SqlUserUpdate, currUser)
				db.Close()
			} else {
				db := database.GetConnection()
				newUser.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
				newUser.ExtID = extId
				db.NamedExec(database.SqlUserInsert, newUser)
				db.Close()
			}
			user, _ := GetUser(r)
			json.NewEncoder(w).Encode(user)
		default:
			http.NotFound(w, r)
		}
	case SafeMatch("/users/([0-1a-zA-Z-]).", r.URL.Path):
		switch r.Method {
		case http.MethodGet:
			user, _ := GetUser(r)
            logger.Printf("%+v", user)
			json.NewEncoder(w).Encode(user)
		default:
			http.NotFound(w, r)
		}
    case SafeMatch("/users/me", r.URL.Path):
        if r.Method == http.MethodGet {
            user, _ := GetUser(r)
            logger.Printf("%+v", user)
            json.NewEncoder(w).Encode(user)
        } else {
            http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
        }
	default:
		http.NotFound(w, r)
	}
}
