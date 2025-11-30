package models

import "github.com/jackc/pgx/v5/pgxpool"

type Models struct {
	Posts    PostsInterface
	Users    UsersInterface
	Comments CommentsInterface
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
	}
}
