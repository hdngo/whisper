package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/hdngo/whisper/internal/model"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *model.Message) error {
	var err error
	for attempts := 0; attempts < 3; attempts++ {
		err = r.createWithTimeout(ctx, msg)
		if err == nil {
			return nil
		}

		if errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, sql.ErrConnDone) ||
			errors.Is(err, sql.ErrTxDone) {
			time.Sleep(time.Duration(attempts+1) * 100 * time.Millisecond)
			continue
		}

		return err
	}
	return err
}

func (r *MessageRepository) createWithTimeout(ctx context.Context, msg *model.Message) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
        INSERT INTO messages (content, user_id, username, created_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

	err := r.db.QueryRowContext(
		ctx,
		query,
		msg.Content,
		msg.UserID,
		msg.Username,
		time.Now().Unix(),
	).Scan(&msg.ID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("no rows affected")
		}
		return err
	}

	return nil
}

func (r *MessageRepository) GetRecent(ctx context.Context, limit int) ([]model.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, content, user_id, username, created_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.Content,
			&msg.UserID,
			&msg.Username,
			&msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse the slice to get chronological order
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - i - 1
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *MessageRepository) GetMessagesBefore(ctx context.Context, beforeID int64, limit int) ([]model.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
        SELECT id, content, user_id, username, created_at
        FROM messages
        WHERE id < $1
        ORDER BY id DESC
        LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, beforeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.Content,
			&msg.UserID,
			&msg.Username,
			&msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse the slice to get chronological order
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - i - 1
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
