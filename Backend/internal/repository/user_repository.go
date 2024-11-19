package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/hdngo/whisper/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id`

	return r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password,
		time.Now().Unix(),
	).Scan(&user.ID)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1`

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
