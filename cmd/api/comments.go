package main

import (
	"errors"
	"net/http"
	"social/internal/models"

	"github.com/google/uuid"
)

// commentPayload represents the payload to create a comment
// swagger:model commentPayload
type commentPayload struct {
	// Content of the comment
	// example: Nice post!
	Content string `json:"content" validate:"required,max=1000" example:"Nice post!"`
	// PostID is injected from path and validated internally
	PostID uuid.UUID `json:"-" validate:"required,uuid"`
}

// createCommentHandler godoc
//
//	@Summary		Create a comment for a post
//	@Description	Creates a new comment associated with the specified post
//	@Tags			Comments
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		string			true	"Post ID (UUID)"
//	@Param			request	body		commentPayload	true	"Comment payload"
//	@Success		201		{object}	DataResponseComment
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/posts/{postID}/comments [post]
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	cp := commentPayload{}
	if err := readJSON(w, r, &cp); err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	post := getPostFromContext(r)
	cp.PostID = post.ID

	if err := Validate.Struct(cp); err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	// todo: get the userID from request context
	userID := uuid.MustParse("b58e1f73-028f-4c17-b8ac-8a3b416c69fd")

	commentUID, err := uuid.NewV7()
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}

	comment := models.Comment{
		ID:      commentUID,
		Content: cp.Content,
		PostID:  cp.PostID,
		UserID:  userID,
	}

	if err := app.models.Comments.CreateComment(r.Context(), &comment); err != nil {
		if errors.Is(err, models.ErrForeignKeyViolation) {
			app.errorBadRequest(w, r, errors.New("post does not exist"))
			return
		}
		app.errorServerError(w, r, err)
		return
	}

	if err = app.jsonResponse(w, http.StatusCreated, &comment); err != nil {
		app.errorServerError(w, r, err)
	}
}
