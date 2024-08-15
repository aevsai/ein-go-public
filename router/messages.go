package router

import (
	"database/sql"
	"einstein-server/chatgpt"
	"einstein-server/database"
	"einstein-server/router/schemes"
	"einstein-server/storage"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func SendErrorMessage(user database.User) {
    db := database.GetConnection()
    answer := database.Message{
        ID:      uuid.New(),
        UserId:  user.ID.UUID,
        Content: "Sorry, service temporary unavailable 🥲",
        Role:    database.RoleAssistant,
    }
    db.NamedExec(database.SqlMessageInsert, answer)
    db.Close()
    logger.Error().Msg("Unable to get message")
}

func asyncCompletionRequest(messages []database.Message, user database.User) {
	db := database.GetConnection()
    defer db.Close()
	_, err := chatgpt.RequestCompletion(user, db, nil)
    
	if err != nil {
        SendErrorMessage(user)
        return
	}
}

func HandleMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
        params := r.URL.Query()
        limit, err := strconv.Atoi(params.Get("limit"))
        if err != nil {
            limit = 20
        }
        offset, err := strconv.Atoi(params.Get("offset"))
        if err != nil {
            offset = 0
        }
		var messages []database.Message
		user, _ := GetUser(r)
        messages, err = database.SelectMessageByUserDisplay(user, &offset, &limit)
        if err != nil {
            logger.Err(err).Msg("")
            http.Error(w, "Error while selecting messages.", http.StatusInternalServerError)
            return
        }
        db := database.GetConnection()
        defer db.Close()
        var usageStat database.UsageStatistics
        err = db.Get(&usageStat, database.SqlUsageStatisticsSelectByUserAndDate, user.ID.UUID, time.Now())
        if err != nil {
            if err != sql.ErrNoRows {
                logger.Err(err).Msg("")
                http.Error(w, "Error while selecting messages.", http.StatusInternalServerError)
                return           
            }
        }
        response := schemes.MessagesResponse{
            Messages: messages,
            Limit: schemes.Limit{
                Total: 5,
                Remained: 5-usageStat.MessagesAmount,
            },
        }
        if err = json.NewEncoder(w).Encode(response); err != nil {
            logger.Err(err).Msg("")
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        }
	case http.MethodPost:
		var newMessage database.Message
		var errAnswer database.Message
		if err := json.NewDecoder(r.Body).Decode(&newMessage); err != nil {
			logger.Err(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
        logger.Debug().Msgf("%+v", newMessage)
		db := database.GetConnection()
        defer db.Close()
		if len(newMessage.Content) > 16385 {
			errAnswer = database.Message{ID: uuid.New(), Content: "Messsage too long, please shorten it. 😅", Role: database.RoleAssistant}
			// set summarizedContent to empty string, so it wont get to chatgpt
			newMessage.SummarizedContent = sql.NullString{String: "", Valid: true}
		}
		newMessage.ID = uuid.New()
		user, _ := GetUser(r)
		newMessage.UserId = user.ID.UUID
        _, err := db.NamedExec(database.SqlMessageInsert, newMessage)
        if err != nil {
            logger.Err(err)
            http.Error(w, "Error while inserting message.", http.StatusInternalServerError)
        }
        if len(newMessage.Attachments) > 0 {
            attachment := newMessage.Attachments[0]
            attachment.MessageID = uuid.NullUUID{UUID: newMessage.ID, Valid: true}
            _, err = db.NamedExec(database.SqlAttachmentUpdate, attachment)
            if err != nil {
                logger.Printf("Error while updating attachment: %s", err)
                http.Error(w, "Error while updating attachment.", http.StatusInternalServerError)
            }
        }
		if len(newMessage.Content) < 16385 {
			var messages []database.Message
            messages, err = database.SelectMessageByUser(user)
            if err != nil {
                logger.Err(err)
                http.Error(w, "Error while selecting messages for chatgpt", http.StatusInternalServerError)
            }
			go asyncCompletionRequest(messages, user)
		} else {
			db.NamedExec(database.SqlMessageInsert, errAnswer)
		}
		db.Close()
		json.NewEncoder(w).Encode(newMessage)
	case http.MethodPut:
		var message database.Message
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			logger.Err(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		user, _ := GetUser(r)
		message.UserId = user.ID.UUID
		db := database.GetConnection()
        defer db.Close()
		_, err := db.NamedExec(database.SqlMessageUpdate, message)
		if err != nil {
			logger.Printf("Error: %s \n", err)
		}
		db.Close()
		json.NewEncoder(w).Encode(message)
	case http.MethodDelete:
		var message database.Message
        var err error
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			logger.Err(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		db := database.GetConnection()
        defer db.Close()
        strg := storage.NewClient()
        for _, v := range message.Attachments {
            _, err = db.Exec(database.SqlAttachmentDelete, v.ID)
            if err != nil {
                logger.Err(err)
                http.Error(w, "Error deleting message.", http.StatusInternalServerError)
            }
            strg.DeleteFile(v.Key)
        }
        _, err = db.Exec(database.SqlMessageDelete, message.ID)
        if err != nil {
            logger.Err(err)
            http.Error(w, "Error deleting message.", http.StatusInternalServerError)
        }
		db.Close()
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.NotFound(w, r)
	}
}
