// Package database
// Author: Evsikov Artem

package database

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	SqlMessageInsert              = `INSERT INTO public.messages (id, user_id, content, smart_content, summarized, role) 
                                     VALUES(:id, :user_id, :content, :smart_content, :summarized, :role);`
	SqlMessageSelect              = `SELECT * FROM messages WHERE id=$1 ORDER BY created_at DESC;`
    SqlMessageUpdate              = `UPDATE public.messages SET user_id=:user_id, content=:content, smart_content=:smart_content, 
    summarized=:summarized, role=:role, in_context=:in_context, in_context_by_force=:in_context_by_force WHERE id=:id;`
	SqlMessageSelectByUserDisplay = `SELECT * FROM public.messages m
                                     WHERE user_id=$1 AND role in ('assistant', 'user')
                                     ORDER BY m.created_at DESC
                                     LIMIT $2 OFFSET $3;`
	SqlMessageSelectByUser        = `SELECT * FROM public.messages WHERE user_id=$1 ORDER BY created_at ASC;`
    SqlMessageDelete              = `DELETE FROM public.messages WHERE id=$1;`
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

const (
    SmartContentTypeButton = "button"
)
const (
    SmartContentKeyGoogleAuth = "google-auth"
)


type SmartContent struct {
    Type string `json:"type"`
    Key string `json:"key"`
    Value   string  `json:"value"`
}

type Message struct {
	ID                uuid.UUID      `json:"id"  db:"id"`
	UserId            uuid.UUID      `json:"user_id"  db:"user_id"`
	Content           string         `json:"content"  db:"content"`
    SmartContent      *json.RawMessage `json:"smart_content" db:"smart_content"`        
	SummarizedContent sql.NullString `json:"summarizedContent" db:"summarized"`
	Role              string         `json:"role"  db:"role"`
	InContext         bool           `json:"in_context" db:"in_context"`
	InContextByForce  bool           `json:"in_context_by_force" db:"in_context_by_force"`
    Attachments       []Attachment   `json:"attachments"`
	CreatedAt         time.Time      `json:"created_at"  db:"created_at"`
}

func SelectMessageByUserDisplay(user User, offset *int, limit *int) ([]Message, error){
    
    if offset == nil {
        defaultOffset := 0
        offset = &defaultOffset
    }

    if limit == nil {
        defaultLimit := 20
        limit = &defaultLimit
    }

    var messages []Message
    db := GetConnection()
    defer db.Close()
    err := db.Select(&messages, SqlMessageSelectByUserDisplay, user.ID, *limit, *offset)
    if err != nil {
        logger.Err(err).Msg("")
        return []Message{}, err
    }
    var attachments []Attachment
    var msgIds []string
    for _, v := range messages {
        msgIds = append(msgIds, v.ID.String())
    }
    if len(msgIds) > 0 {
        query, args, err := sqlx.In(SqlAttachmentSelectByMessageIds, msgIds)
        if err != nil {
            logger.Err(err).Msg("")
            return []Message{}, err
        }
        query = db.Rebind(query)

        err = db.Select(&attachments, query, args...)
        if err != nil {
            logger.Err(err).Msg("")
            return []Message{}, err
        }
        attchMap := make(map[uuid.UUID][]Attachment)
        for _, v := range attachments {
            attchMap[v.MessageID.UUID] = append(attchMap[v.MessageID.UUID], v)
        }
        for i := range messages {
            messages[i].Attachments = attchMap[messages[i].ID]
        }
    }
    db.Close()
    return messages, nil
}

func SelectMessageByUser(user User) ([]Message, error){
    var messages []Message
    db := GetConnection()
    defer db.Close()
    err := db.Select(&messages, SqlMessageSelectByUser, user.ID)
    if err != nil {
        return []Message{}, err
    }
    var attachments []Attachment
    var msgIds []string
    for _, v := range messages {
        msgIds = append(msgIds, v.ID.String())
    }
    query, args, err := sqlx.In(SqlAttachmentSelectByMessageIds, msgIds)
    if err != nil {
        return []Message{}, err
    }
    query = db.Rebind(query)
    
    err = db.Select(&attachments, query, args...)
    if err != nil {
        return []Message{}, err
    }
    attchMap := make(map[uuid.UUID][]Attachment)
    for _, v := range attachments {
        attchMap[v.MessageID.UUID] = append(attchMap[v.MessageID.UUID], v)
    }
    for i := range messages {
        messages[i].Attachments = attchMap[messages[i].ID]
    }
    db.Close()
    return messages, nil
}

// NullRawMessage represents a json.RawMessage that may be null.
// NullRawMessage implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullRawMessage struct {
	RawMessage json.RawMessage
	Valid bool // Valid is true if JSON is not NULL
}

// Scan implements the Scanner interface.
func (n *NullRawMessage) Scan(value interface{}) error {
	if value == nil {
		n.RawMessage, n.Valid = json.RawMessage{}, false
		return nil
	}
	buf, ok := value.([]byte)

	if !ok {
		return fmt.Errorf("canot parse to bytes")
	}

	n.RawMessage, n.Valid = buf, true

	return nil
}

// Value implements the driver Valuer interface.
func (n NullRawMessage) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.RawMessage.MarshalJSON()
}
