package main

import (
	"net/http"
	"social/internal/models"

	"github.com/google/uuid"
)

type postPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload postPayload
	err := readJSON(w, r, &payload)
	if err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	err = Validate.Struct(payload)
	if err != nil {
		app.errorBadRequest(w, r, err)
		return
	}
	postID, err := uuid.NewV7()
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}
	// todo: get userID from request context
	userID := uuid.MustParse("b58e1f73-028f-4c17-b8ac-8a3b416c69fd")

	post := &models.Post{
		ID:      postID,
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  userID,
	}
	err = app.models.Posts.Create(r.Context(), post)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}

	err = app.jsonResponse(w, http.StatusCreated, post)
	if err != nil {
		app.errorServerError(w, r, err)
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)
	comments, err := app.models.Comments.GetComments(r.Context(), post.ID)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}
	post.Comments = comments

	err = app.jsonResponse(w, http.StatusOK, post)
	if err != nil {
		app.errorServerError(w, r, err)
	}
}
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)
	err := app.models.Posts.Delete(r.Context(), post.ID)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}

	err = app.jsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		app.errorServerError(w, r, err)
	}
}

type updatePostPayload struct {
	Title   *string  `json:"title" validate:"omitempty,max=100"`
	Content *string  `json:"content" validate:"omitempty,max=1000"`
	Tags    []string `json:"tags" validate:"omitempty"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)

	var payload updatePostPayload
	err := readJSON(w, r, &payload)
	if err != nil {
		app.errorBadRequest(w, r, err)
	}

	err = Validate.Struct(payload)
	if err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Tags != nil {
		post.Tags = payload.Tags
	}

	err = app.models.Posts.Update(r.Context(), post)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}

	err = app.jsonResponse(w, http.StatusOK, post)
	if err != nil {
		app.errorServerError(w, r, err)
	}
}
