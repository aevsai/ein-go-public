package schemes

import "einstein-server/database"

type Limit struct {
    Total int `json:"total"`
    Remained int `json:"remained"`
}

type MessagesResponse struct {
    Messages []database.Message `json:"messages"`
    Limit Limit `json:"limit"`
}
