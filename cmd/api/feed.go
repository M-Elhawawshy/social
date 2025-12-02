package main

import (
	"net/http"
	"social/internal/models"

	"github.com/google/uuid"
)

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
