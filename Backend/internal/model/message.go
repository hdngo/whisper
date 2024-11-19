package model

type Message struct {
	ID        int64  `json:"id" db:"id"`
	Content   string `json:"content" db:"content"`
	UserID    int64  `json:"user_id" db:"user_id"`
	Username  string `json:"username" db:"username"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

const (
	MessageTypeChat  = "chat"
	MessageTypeJoin  = "join"
	MessageTypeLeave = "leave"
	MessageTypeUsers = "users"
)
