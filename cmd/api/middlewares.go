package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"social/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type postKey string
type userKey string

const (
	postCtxKey postKey = "post"
	userCTXKey userKey = "user"
)

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := uuid.Parse(chi.URLParam(r, "postID"))
		if err != nil {
			app.errorBadRequest(w, r, err)
			return
		}
		post, err := app.models.Posts.GetByID(r.Context(), postID)
		if err != nil {
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				app.errorNotFound(w, r, err)
				return
			default:
				app.errorServerError(w, r, err)
				return
			}
		}
		ctx := context.WithValue(r.Context(), postCtxKey, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromContext(r *http.Request) *models.Post {
	post, _ := r.Context().Value(postCtxKey).(*models.Post)
	return post
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			app.errorBadRequest(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.models.Users.Get(ctx, userUUID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				app.errorNotFound(w, r, err)
				return
			default:
				app.errorServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, userCTXKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *models.User {
	return r.Context().Value(userCTXKey).(*models.User)
}
