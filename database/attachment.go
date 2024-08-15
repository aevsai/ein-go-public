package database

import (
	"time"

	"github.com/google/uuid"
)

const (
	SqlAttachmentSelect        = `SELECT id, message_id, bucket, "key", created_at FROM public.attachments;`
    SqlAttachmentInsert        = `INSERT INTO public.attachments (id, message_id, bucket, "key", created_at) VALUES(:id, :message_id, :bucket, :key, now());`
    SqlAttachmentUpdate        = `UPDATE public.attachments SET message_id=:message_id WHERE id=:id;`
	SqlAttachmentSelectByMessageId = `SELECT id, message_id, bucket, "key", "created_at" FROM public.attachments WHERE message_id=$1`
	SqlAttachmentSelectById = `SELECT id, message_id, bucket, "key", "created_at" FROM public.attachments WHERE id=$1`
	SqlAttachmentSelectByMessageIds = `SELECT id, message_id, bucket, "key", "created_at" FROM public.attachments WHERE message_id in (?)`
    SqlAttachmentDelete        = `DELETE FROM public.attachments WHERE id=$1`
)

type Attachment struct {
    ID          uuid.UUID           `json:"id" db:"id"`
    MessageID   uuid.NullUUID       `json:"message_id" db:"message_id"`
    Bucket      string              `json:"bucket" db:"bucket"`
    Key         string              `json:"key" db:"key"`
    CreatedAt   time.Time           `json:"created_at" db:"created_at"`
}
