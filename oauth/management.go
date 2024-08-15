package oauth

import (
	"context"
    "einstein-server/database"
    "fmt"
    "os"
    "github.com/auth0/go-auth0/management"
    "github.com/google/uuid"
    "golang.org/x/oauth2"
)

func GetClient() (*management.Management, error) {
    // Initialize a new client using a domain, client ID and secret.
    m, err := management.New(
        os.Getenv("AUTH0_DOMAIN"),
        management.WithClientCredentials(
            context.TODO(), 
            os.Getenv("AUTH0_CLIENT_ID"), 
            os.Getenv("AUTH0_CLIENT_SECRET"),
        ),
    )
    if err != nil {
        logger.Err(err).Msg("")
        return nil, err
    }
    return m, nil
}

func GetUserIdPToken(userId uuid.UUID, provider string) (oauth2.Token, error) {
    client, err := GetClient()
    tok := oauth2.Token{} 

    if err != nil {
        logger.Err(err).Msg("")
        return tok, err
    }
    db := database.GetConnection()
    var dbUser database.User
    if err = db.Get(&dbUser, database.SqlUserSelect, userId); err != nil {
        logger.Err(err).Msg("")
        return tok, err
    }
    users, err := client.User.ListByEmail(context.TODO(), dbUser.Email)
    if len(users) > 0 {
        for _, user := range users { 
            for _, i := range user.Identities {
                if *i.Provider == provider {
                    if i.AccessToken != nil {
                        tok.AccessToken = *i.AccessToken
                    }
                    if i.RefreshToken != nil {
                        tok.RefreshToken = *i.RefreshToken
                    }
                    return tok, nil 
                }
            }
        }
    } else {
        err = fmt.Errorf("There is no user with such email.")
        logger.Err(err).Msg("")
        return tok, err
    }
    return tok, nil
}

func AttachSecondaryUser(emailPrimary, emailSecondary string) error {
    client, err := GetClient()
    if err != nil {
        logger.Err(err).Msg("")
        return err
    }
    users, err := client.User.ListByEmail(context.TODO(), emailPrimary)
    primaryUser := users[0]
    users, err = client.User.ListByEmail(context.TODO(), emailSecondary)
    secondaryUser := users[0]
    secondaryIdentity := getGoogleIdentity(*secondaryUser)
    client.User.Link(context.TODO(), *primaryUser.ID, &management.UserIdentityLink{
        UserID: secondaryUser.ID,
        Provider: secondaryIdentity.Provider,
    })
    return nil
}

func getGoogleIdentity(user management.User) *management.UserIdentity {
    for _, i := range user.Identities {
        if *i.Provider == "google-oauth2" {
            return i
        }
    }
    return nil
}
