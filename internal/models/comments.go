package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentsInterface interface {
	GetComments(context.Context, uuid.UUID) ([]Comment, error)
	CreateComment(context.Context, *Comment) error
}

type Comment struct {
	ID        uuid.UUID `json:"id"`
	Content   string    `json:"content"`
	PostID    uuid.UUID `json:"post_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user"`
}

type CommentsModel struct {
	pool *pgxpool.Pool
}

func (c *CommentsModel) CreateComment(ctx context.Context, comment *Comment) error {
	statement := `
		INSERT INTO comments(ID, CONTENT, POST_ID, USER_ID)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	err := c.pool.QueryRow(ctx, statement, comment.ID, comment.Content, comment.PostID, comment.UserID).Scan(&comment.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				return ErrForeignKeyViolation
			}
		}
		return err
	}

	return nil
}

func (c *CommentsModel) GetComments(ctx context.Context, postID uuid.UUID) ([]Comment, error) {
	statement := `
		SELECT c.id, c.content, c.post_id, c.user_id, c.created_at, u.id, u.username, u.email
		FROM comments c
    		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC;`

	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	rows, err := c.pool.Query(ctx, statement, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var comment Comment
		comment.User = User{}
		err = rows.Scan(&comment.ID, &comment.Content, &comment.PostID, &comment.UserID, &comment.CreatedAt, &comment.User.ID, &comment.User.Username, &comment.User.Email)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
