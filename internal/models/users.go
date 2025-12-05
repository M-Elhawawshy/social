package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersInterface interface {
	Create(context.Context, *User) error
	Get(context.Context, uuid.UUID) (*User, error)
	Update(context.Context, *User) error
	Follow(context.Context, uuid.UUID, uuid.UUID) error
	Unfollow(context.Context, uuid.UUID, uuid.UUID) error
	CreateUserAndInvite(context.Context, *User) error
}

type User struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	IsActivated bool      `json:"is_activated"`
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
	var passwordBytes []byte
	err := u.pool.QueryRow(ctx, statement, userID).Scan(&user.ID, &user.Username, &user.Email, &passwordBytes, &user.CreatedAt)
	user.Password = string(passwordBytes)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserModel) Update(ctx context.Context, user *User) error {
	statement := `
		UPDATE users u 
		SET username = $1, is_activated = $2, email = $3
		WHERE u.id = $4
`

	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	_, err := u.pool.Exec(ctx, statement, user.Username, user.IsActivated, user.Email, user.ID)
	return err
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

func (u *UserModel) CreateUserAndInvite(ctx context.Context, user *User) error {
	return executeWithTx(u.pool, ctx, func(tx pgx.Tx) error {
		err := createUserTx(ctx, tx, user)
		if err != nil {
			return err
		}
		invite := &Invite{
			UserID:      user.ID,
			InviteToken: uuid.New(),
			ExpiresAt:   time.Now().Add(InviteExpirationTime),
		}
		return createInvite(ctx, tx, invite)
	})
}

func createUserTx(ctx context.Context, tx pgx.Tx, user *User) error {
	statement := `
			INSERT INTO users (id, username, email, password)
			VALUES ($1, $2, $3, $4)
			RETURNING is_activated, created_at
		`

	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	return tx.QueryRow(ctx, statement, user.ID, user.Username, user.Email, user.Password).Scan(&user.IsActivated, &user.CreatedAt)
}
