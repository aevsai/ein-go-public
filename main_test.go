package chatgpt

import (
	"einstein-server/database"
	"testing"

	"github.com/google/uuid"
)

func TestRequestCompletion(t *testing.T) {
	db := database.GetConnection()

	// Create a user to test
	user := database.User{ID: uuid.NullUUID{}, ExtID: uuid.NewString(), Name: "John Doe", Email: "johndoe@example.org", ProfilePicture: "pic.jpg"}
	user.ID.UUID = uuid.New()
	user.ID.Valid = true
	_, err := db.NamedExec(database.SqlUserInsert, user)
	if err != nil {
		t.Fatalf("Error while inserting into database: %+v", err)
	}

	var createdUser database.User
	err = db.Get(&createdUser, database.SqlUserSelect, user.ID)
	if err != nil {
		t.Fatalf("Error while reading from database: %+v", err)
	}

	// Create a messages to test
	messages := []database.Message{
		{ID: uuid.New(), UserId: createdUser.ID.UUID, Content: "Hello", Role: database.RoleUser},
	}

	for _, v := range messages {
		_, err = db.NamedExec(database.SqlMessageInsert, v)
	}

	if err != nil {
		t.Fatalf("Error while inserting into database: %+v", err)
	}

	db.Select(&messages, database.SqlMessageSelectByUser, user.ID)

	_, err = RequestCompletion(user, db, nil)

	if err != nil {
		t.Fatalf("Error while requesting completion %s", err)
	}

	if len(messages) == 0 {
		t.Fatalf("Messages for user %+v were not retrieved", user.ID)
	}
}

func TestRequestCompletionWithFunctions(t *testing.T) {
	db := database.GetConnection()

	// Create a user to test
	user := database.User{ID: uuid.NullUUID{}, ExtID: uuid.NewString(), Name: "John Doe", Email: "johndoe@example.org", ProfilePicture: "pic.jpg"}
	user.ID.UUID = uuid.New()
	user.ID.Valid = true
	_, err := db.NamedExec(database.SqlUserInsert, user)
	if err != nil {
		t.Fatalf("Error while inserting into database: %+v", err)
	}

	var createdUser database.User
	err = db.Get(&createdUser, database.SqlUserSelect, user.ID)
	if err != nil {
		t.Fatalf("Error while reading from database: %+v", err)
	}

	// Create a messages to test
	messages := []database.Message{
		{ID: uuid.New(), UserId: createdUser.ID.UUID, Content: "What weather in Moscow", Role: database.RoleUser},
	}

	for _, v := range messages {
		_, err = db.NamedExec(database.SqlMessageInsert, v)
	}

	if err != nil {
		t.Fatalf("Error while inserting into database: %+v", err)
	}

	db.Select(&messages, database.SqlMessageSelectByUser, user.ID)

	_, err = RequestCompletion(user, db, nil)

	if err != nil {
		t.Fatalf("Error while requesting completion %s", err)
	}

	if len(messages) == 0 {
		t.Fatalf("Messages for user %+v were not retrieved", user.ID)
	}
}
