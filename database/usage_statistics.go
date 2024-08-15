package database

import (
	"time"

	"github.com/google/uuid"
)

const (
    SqlUsageStatisticsInsert              = `INSERT INTO public.usage_statistics (id, user_id, date, tokens_in, tokens_out, messages_amount) VALUES(:id, :user_id, :date, :tokens_in, :tokens_out, :messages_amount);`
	SqlUsageStatisticsSelect              = `SELECT * FROM public.usage_statistics WHERE id=$1 ORDER BY date DESC;`
    SqlUsageStatisticsUpdate              = `UPDATE public.usage_statistics SET user_id=:user_id, tokens_in=:tokens_in, tokens_out=:tokens_out, messages_amount=:messages_amount WHERE user_id=:user_id AND date=:date;`
	SqlUsageStatisticsSelectByUser        = `SELECT * FROM public.usage_statistics WHERE user_id=$1 ORDER BY date DESC;`
	SqlUsageStatisticsSelectByUserAndDate = `SELECT * FROM public.usage_statistics WHERE user_id=$1 AND date=$2 ORDER BY date DESC;`
)

type UsageStatistics struct {
	ID              uuid.UUID `json:"id"  db:"id"`
	UserId          uuid.UUID `json:"user_id"  db:"user_id"`
	Date            time.Time `json:"date"  db:"date"`
	TokensIn        int       `json:"tokens_in"  db:"tokens_in"`
	TokensOut       int       `json:"tokens_out"  db:"tokens_out"`
    MessagesAmount  int       `json:"messages_amount" db:"messages_amount"`
}
