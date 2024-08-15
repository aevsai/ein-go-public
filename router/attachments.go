package router

import (
	"bytes"
	"einstein-server/database"
	"einstein-server/storage"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func HandleUploadFile(w http.ResponseWriter, r *http.Request) {

    switch r.Method {
    case http.MethodPost:
        // Parse the multipart form
        if err := r.ParseMultipartForm(10 << 20); err != nil {
            logger.Err(err)
            http.Error(w, "Error parsing the form.", http.StatusBadRequest)
        }

        // Get the file from the form
        file, handler, err := r.FormFile("file")
        if err != nil {
            logger.Err(err)
            http.Error(w, "Error retrieving the file.", http.StatusInternalServerError)
        }
        defer file.Close()

        strg := storage.NewClient()
        buf := bytes.NewBuffer(nil)
        if _, err := io.Copy(buf, file); err != nil {
            logger.Err(err)
            http.Error(w, "Error while reading file.", http.StatusInternalServerError)
        }
        name := time.Now().Format("02 Jan 06 15:04 MST")+" - "+handler.Filename
        err = strg.UploadFile(buf.Bytes(), name)
        if err != nil {
            logger.Err(err)
            http.Error(w, "Error creating the file", http.StatusInternalServerError)
        }

        logger.Debug().Msgf("File uploaded successfully: %s", handler.Filename)
        attachment := database.Attachment{
            ID: uuid.New(),
            Key: name,
        }
        db := database.GetConnection()
        defer db.Close()
        _, err = db.NamedExec(database.SqlAttachmentInsert, attachment)
        if err != nil {
            logger.Err(err)
            http.Error(w, "Error while inserting attachment.", http.StatusInternalServerError)
        }
        err = db.Get(&attachment, database.SqlAttachmentSelectById, attachment.ID)
        if err != nil {
            logger.Err(err)
            http.Error(w, "Error selecting created attachment.", http.StatusInternalServerError)
        }
        json.NewEncoder(w).Encode(attachment)
    case http.MethodPut:
        var newAttachment database.Attachment
        var oldAttachment database.Attachment
        if err := json.NewDecoder(r.Body).Decode(&newAttachment); err != nil {
            logger.Err(err)
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
        db := database.GetConnection()
        defer db.Close()
        err := db.Get(&oldAttachment, database.SqlAttachmentSelect, newAttachment.ID)
        if err != nil {
            logger.Err(err)
            http.Error(w, "Error requesting attachment.", http.StatusInternalServerError)
        }
        oldAttachment.MessageID = newAttachment.MessageID
        _, err = db.NamedExec(database.SqlAttachmentUpdate, oldAttachment)
        if err != nil {
            logger.Err(err)
            http.Error(w, "Error updating attachment.", http.StatusInternalServerError)
        }
        json.NewEncoder(w).Encode(oldAttachment)
    default:
        http.NotFound(w, r)
    }
}
