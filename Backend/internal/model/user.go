package model

type User struct {
	ID        int64  `json:"id" db:"id"`
	Username  string `json:"username" db:"username"`
	Password  string `json:"-" db:"password_hash"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}
