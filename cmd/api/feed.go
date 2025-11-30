package main

import (
	"net/http"

	"github.com/google/uuid"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// todo: replace this with user from context
	userID := uuid.MustParse("bd3319ea-d197-46cf-9b79-4a03e136c22d")

	posts, err := app.models.Posts.Feed(r.Context(), userID)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}

	if err = app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.errorServerError(w, r, err)
	}
}
