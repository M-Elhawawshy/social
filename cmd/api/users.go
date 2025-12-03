package main

import (
	"errors"
	"net/http"
	"social/internal/models"

	"github.com/google/uuid"
)

// getUserHandler godoc
//
//	@Summary		Get a user
//	@Description	Retrieves a user by ID
//	@Tags			Users
//	@Produce		json
//	@Param			userID	path		string	true	"User ID (UUID)"
//	@Success		200		{object}	DataResponseUser
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/users/{userID} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.errorServerError(w, r, err)
	}
}

// followUserHandler godoc
//
//	@Summary		Follow a user
//	@Description	Follow the specified user
//	@Tags			Users
//	@Param			userID	path	string	true	"User ID (UUID)"
//	@Success		204		"No Content"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/users/{userID}/follow [put]
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

// unfollowUserHandler godoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow the specified user
//	@Tags			Users
//	@Param			userID	path	string	true	"User ID (UUID)"
//	@Success		204		"No Content"
//	@Failure		500		{object}	ErrorResponse
//	@Router			/users/{userID}/unfollow [put]
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
