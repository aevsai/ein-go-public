package database

import (
	"github.com/google/uuid"
	"time"
)

const (
	SqlToolInsert       = `INSERT INTO public.messages (id, user_id, content, summarized, role) VALUES(:id, :user_id, :content, :summarized, :role);`
	SqlToolSelect       = `SELECT * FROM messages WHERE id=$1 ORDER BY created_at DESC;`
	SqlToolUpdate       = `UPDATE public.messages SET user_id=:user_id, "content"=:content, "summarized"=:summarized, "role"=:role WHERE id=:id;`
	SqlToolSelectByUser = `SELECT * FROM messages WHERE user_id=$1 ORDER BY created_at DESC;`
)

type Tool struct {
	ID        uuid.UUID `json:"id"  db:"id"`
	Name      string    `json:"name"  db:"name"`
	Host      string    `json:"host"  db:"host"`
	Method    string    `json:"method" db:"method"`
	Params    []byte    `json:"params"  db:"params"`
	CreatedAt time.Time `json:"created_at"  db:"created_at"`
}
