package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostsInterface interface {
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, post *Post) error
	Feed(ctx context.Context, userID uuid.UUID) ([]FeedPost, error)
}

type Post struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type FeedPost struct {
	Post
	CommentsCount     int       `json:"comments_count"`
	TopCommentContent string    `json:"top_comment_content"`
	TopCommentUserID  uuid.UUID `json:"top_comment_user_id"`
}

type PostsModel struct {
	pool *pgxpool.Pool
}

func (p *PostsModel) Create(ctx context.Context, post *Post) error {
	statement := `
			INSERT INTO posts (id, title, content, user_id, tags)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING created_at, updated_at
		`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()
	err := p.pool.QueryRow(ctx, statement, post.ID, post.Title, post.Content, post.UserID, post.Tags).Scan(&post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostsModel) GetByID(ctx context.Context, id uuid.UUID) (*Post, error) {
	statement := `
		SELECT posts.id, posts.title, posts.content, posts.tags, posts.user_id, posts.created_at, posts.updated_at , posts.version
		FROM posts
		WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()
	post := &Post{}
	err := p.pool.QueryRow(ctx, statement, id).Scan(&post.ID, &post.Title, &post.Content, &post.Tags, &post.UserID, &post.CreatedAt, &post.UpdatedAt, &post.Version)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (p *PostsModel) Delete(ctx context.Context, id uuid.UUID) error {
	statement := `DELETE FROM posts WHERE id = $1`
	_, err := p.pool.Exec(ctx, statement, id)
	return err
}

func (p *PostsModel) Update(ctx context.Context, post *Post) error {
	statement := `
		UPDATE posts
		SET title = $1, content = $2, tags = $3, version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()
	err := p.pool.QueryRow(ctx, statement, post.Title, post.Content, post.Tags, post.ID, post.Version).Scan(&post.Version)
	return err
}

func (p *PostsModel) Feed(ctx context.Context, userID uuid.UUID) ([]FeedPost, error) {
	statement := `
		SELECT p.id, p.title, p.content, p.user_id, p.tags, p.created_at,
       (SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS comments_count,
       u.username, u.id,
       (SELECT content FROM comments WHERE post_id = p.id ORDER BY created_at DESC limit 1) AS top_comment,
       (SELECT user_id FROM comments WHERE post_id = p.id ORDER BY created_at DESC LIMIT 1) AS top_comment_user_id
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1 OR p.user_id IN (SELECT user_id from followers WHERE follower_id = $1)
		ORDER BY p.created_at DESC
		LIMIT 50 offset 0;
	`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	rows, err := p.pool.Query(ctx, statement, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []FeedPost
	for rows.Next() {
		var feedPost FeedPost
		err = rows.Scan(
			&feedPost.ID,
			&feedPost.Title,
			&feedPost.Content,
			&feedPost.UserID,
			&feedPost.Tags,
			&feedPost.CreatedAt,
			&feedPost.CommentsCount,
			&feedPost.User.Username,
			&feedPost.User.ID,
			&feedPost.TopCommentContent,
			&feedPost.TopCommentUserID,
		)
		if err != nil {
			return nil, err
		}
		feed = append(feed, feedPost)
	}
	return feed, nil
}
