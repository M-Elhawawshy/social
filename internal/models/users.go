package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersInterface interface {
	Create(context.Context, *User) error
	Get(context.Context, uuid.UUID) (*User, error)
	Follow(context.Context, uuid.UUID, uuid.UUID) error
	Unfollow(context.Context, uuid.UUID, uuid.UUID) error
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type UserModel struct {
	pool *pgxpool.Pool
}

func (u *UserModel) Create(ctx context.Context, user *User) error {
	statement := `
			INSERT INTO users (id, username, email, password)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at
		`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()
	err := u.pool.QueryRow(ctx, statement, user.ID, user.Username, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt)
	return err
}

func (u *UserModel) Get(ctx context.Context, userID uuid.UUID) (*User, error) {
	statement := `
		SELECT users.ID, USERNAME, EMAIL, PASSWORD, CREATED_AT
		FROM users
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	var user User
	// todo: properly handle password fetching
	var dummy []byte
	err := u.pool.QueryRow(ctx, statement, userID).Scan(&user.ID, &user.Username, &user.Email, &dummy, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserModel) Follow(ctx context.Context, userID uuid.UUID, followerID uuid.UUID) error {
	statement := `INSERT INTO followers(user_id, follower_id) VALUES ($1, $2)`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	_, err := u.pool.Exec(ctx, statement, userID, followerID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" || pgErr.Code == "23505" {
				return ErrForeignKeyViolation
			}
		}
		return err
	}

	return nil
}

func (u *UserModel) Unfollow(ctx context.Context, userID uuid.UUID, followerID uuid.UUID) error {
	statement := `DELETE from followers WHERE user_id = $1 AND follower_id = $2`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	_, err := u.pool.Exec(ctx, statement, userID, followerID)
	return err
}
