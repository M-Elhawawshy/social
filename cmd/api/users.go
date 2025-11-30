package main

import (
	"errors"
	"net/http"
	"social/internal/models"

	"github.com/google/uuid"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.errorServerError(w, r, err)
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	// follow request -> user_id in request + target_id -> simply insert into db
	// todo: fetch userID from context when auth is implemented
	userID := "b58e1f73-028f-4c17-b8ac-8a3b416c69fd"
	userUUID := uuid.MustParse(userID)
	user := getUserFromContext(r)

	err := app.models.Users.Follow(r.Context(), user.ID, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrForeignKeyViolation):
			app.errorBadRequest(w, r, errors.New("already followed or user you are trying to follow does not exist"))
			return
		default:
			app.errorServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.errorServerError(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := "b58e1f73-028f-4c17-b8ac-8a3b416c69fd"
	userUUID := uuid.MustParse(userID)
	user := getUserFromContext(r)

	err := app.models.Users.Unfollow(r.Context(), user.ID, userUUID)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.errorServerError(w, r, err)
	}
}
