// Package database
// Author: Evsikov Artem

package database

import (
	"time"

	"github.com/google/uuid"
)

const (
    SqlUserInsert        = `INSERT INTO public.users (id, ext_id, email, full_name, profilepicture, subscribed, parent_id) 
                            VALUES(:id, :ext_id, :email, :full_name, :profilepicture, :subscribed, :parent_id);`
	SqlUserSelect        = `select * from users where id=$1`
    SqlUserUpdate        = `UPDATE public.users SET email=:email, full_name=:full_name, profilepicture=:profilepicture,
                            subscribed=:subscribed, parent_id=:parent_id WHERE id=:id;`
	SqlUserSelectByExtId = `select * from users where ext_id=$1`
    SqlUserSelectByParentId = `select * from users where parent_id=$1`
    SqlUserSelectAll        = `select * from users`
)

type User struct {
	ID             uuid.NullUUID `json:"id" db:"id"`
	ExtID          string        `json:"ext_id" db:"ext_id"`
	Name           string        `json:"name" db:"full_name"`
	Email          string        `json:"email" db:"email"`
	ProfilePicture string        `json:"profilePicture" db:"profilepicture"`
    Subscribed     bool          `json:"subscribed" db:"subscribed"`
    ParentID       uuid.NullUUID    `json:"parent_id" db:"parent_id"`
    CreatedAt      time.Time     `json:"created_at" db:"created_at"`
}

