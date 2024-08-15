package database

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
    SqlCredentialsInsert = `INSERT INTO public.credentials (id, "type", "data", updated_at, created_at, user_id, service) VALUES(:id, :type, :data, now(), now(), :user_id, :service);`
    SqlCredentialsUpdate = `UPDATE public.credentials SET "type"=:type, "data"=:data, updated_at=now(), user_id=:user_id, service=:service WHERE id=:id;`
    SqlCredentialsSelectByUserAndService = `SELECT * FROM public.credentials WHERE user_id=$1 AND service=$2`
)

type Credentials struct {
    ID uuid.UUID            `json:"id" db:"id"`
    UserID uuid.UUID        `json:"user_id" db:"user_id"`
    Service string          `json:"service" db:"service"`
    Type string             `json:"type" db:"type"`
    Data json.RawMessage    `json:"data" db:"data"`
    UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
    CreatedAt time.Time    `json:"created_at" db:"created_at"`
}
