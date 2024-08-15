package database

import (
	//	"encoding/json"
//	"testing"

//	"github.com/google/uuid"
)

// TestCrudUser calls create-select-update flow for user entity
//func TestCrudUser(t *testing.T) {
//	user := User{ID: uuid.NullUUID{}, ExtID: uuid.NewString(), Name: "John Doe", Email: "johndoe@example.org", ProfilePicture: "pic.jpg"}
//	user.ID.UUID = uuid.New()
//	user.ID.Valid = true
//	db := GetConnection()
//	_, err := db.NamedExec(SqlUserInsert, user)
//	if err != nil {
//		t.Fatalf("Error while inserting into database: %+v", err)
//	}
//
//	var createdUser User
//	err = db.Get(&createdUser, SqlUserSelect, user.ID)
//	if err != nil {
//		t.Fatalf("Error while reading from database: %+v", err)
//	}

//	if createdUser != user {
//		t.Fatalf("Data was not saved correctly: %+v", createdUser)
//	}

//	user.Email = "newjohndoe@example.org"
//	user.Name = "John New Doe"
//	user.ProfilePicture = "newpic.jpg"
//
//	db.NamedExec(SqlUserUpdate, user)
//	var updatedUser User
//	db.Get(&updatedUser, SqlUserSelect, user.ID)
//	if updatedUser != user {
//		t.Fatalf("Data was not updated correctly: %+v", updatedUser)
//	}
//}
