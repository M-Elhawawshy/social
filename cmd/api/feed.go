package main

import (
	"net/http"
	"social/internal/models"

	"github.com/google/uuid"
)

// getUserFeedHandler godoc
//
//	@Summary		Get user feed
//	@Description	Returns a paginated feed of posts from the user and followed users
//	@Tags			Feed
//	@Produce		json
//	@Param			limit	query		int			false	"Items per page"		minimum(1)	maximum(20)
//	@Param			offset	query		int			false	"Offset for pagination"	minimum(0)
//	@Param			sort	query		string		false	"Sort order"			Enums(ASC,DESC)
//	@Param			tags	query		[]string	false	"Filter by tags (comma separated)"
//	@Param			search	query		string		false	"Search in title/content"
//	@Param			from	query		string		false	"From date RFC3339"	example("2024-01-02T15:04:05Z")
//	@Param			to		query		string		false	"To date RFC3339"	example("2024-12-31T23:59:59Z")
//	@Success		200		{object}	DataResponseFeed
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// todo: replace this with user from context
	userID := uuid.MustParse("bd3319ea-d197-46cf-9b79-4a03e136c22d")

	pg, err := models.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "DESC",
	}.Parse(r)
	if err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	err = Validate.Struct(&pg)
	if err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	posts, err := app.models.Posts.Feed(r.Context(), userID, pg)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}

	if err = app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.errorServerError(w, r, err)
	}
}
