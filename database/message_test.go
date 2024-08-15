package database

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestCrudMessage calls create-select-update flow for message entity
func TestCrudMessage(t *testing.T) {
	user := User{ID: uuid.NullUUID{}, ExtID: uuid.NewString(), Name: "John Doe", Email: "johndoe@example.org", ProfilePicture: "pic.jpg"}
	user.ID.UUID = uuid.New()
	user.ID.Valid = true
	db := GetConnection()
	_, err := db.NamedExec(SqlUserInsert, user)
	if err != nil {
		t.Fatalf("Error while inserting into database: %+v", err)
	}

	var createdUser User
	err = db.Get(&createdUser, SqlUserSelect, user.ID)
	if err != nil {
		t.Fatalf("Error while reading from database: %+v", err)
	}

	message := Message{ID: uuid.New(), UserId: createdUser.ID.UUID, Content: "Hello World", Role: RoleAssistant, CreatedAt: time.Now(), InContext: true}
	_, err = db.NamedExec(SqlMessageInsert, message)
	if err != nil {
		t.Fatalf("Error while inserting into database: %+v", err)
	}

	var createdMessage Message
	err = db.Get(&createdMessage, SqlMessageSelect, message.ID)
	if err != nil {
		t.Fatalf("Error while reading from database: %+v", err)
	}
	createdMessage.CreatedAt = message.CreatedAt
	if createdMessage.ID != message.ID || createdMessage.UserId != message.UserId || createdMessage.Content != message.Content {
		t.Fatalf("Data was not saved correctly: %+v, %+v", createdMessage, message)
	}

	message.Content = "Well, hello"
	message.Role = RoleUser

	db.NamedExec(SqlMessageUpdate, message)
	var updatedMessage Message
	db.Get(&updatedMessage, SqlMessageSelect, message.ID)
	updatedMessage.CreatedAt = message.CreatedAt
//	if updatedMessage != message {
//		t.Fatalf("Data was not updated correctly: %+v, %+v", updatedMessage, message)
//	}

	var messages []Message
	db.Select(&messages, SqlMessageSelectByUser, user.ID)

	if len(messages) == 0 {
		t.Fatalf("Messages for user %+v were not retrieved", user.ID)
	}

    messages, err = SelectMessageByUserDisplay(user, nil , nil)
	if err != nil {
        t.Fatalf("Messages for user %+v were not retrieved: %s", user.ID, err)
	}
}
