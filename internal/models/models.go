package models

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	Posts    PostsInterface
	Users    UsersInterface
	Comments CommentsInterface
	Invites  InvitesInterface
}

func NewModels(pool *pgxpool.Pool) *Models {
	return &Models{
		Posts: &PostsModel{
			pool: pool,
		},
		Users: &UserModel{
			pool: pool,
		},
		Comments: &CommentsModel{
			pool: pool,
		},
		Invites: &InvitesModel{
			pool: pool,
		},
	}
}

func executeWithTx(db *pgxpool.Pool, ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err = fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
