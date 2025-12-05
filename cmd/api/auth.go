package main

import (
	"errors"
	"net/http"
	"social/internal/models"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type signupPayload struct {
	Username string `json:"username" validate:"required,max=20,min=3"`
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"min=5,max=20"`
}

// signupHandler godoc
//
//	@Summary		Register a user
//	@Description	Register a new user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		signupPayload	true	"Signup payload"
//	@Success		201		{object}	DataResponseUser
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/auth/signup [post]
func (app *application) signupHandler(w http.ResponseWriter, r *http.Request) {
	var form signupPayload

	if err := readJSON(w, r, &form); err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(form); err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	userID, err := uuid.NewV7()
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}
	hashedPassword, err := hashPassword(form.Password)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}
	user := &models.User{
		ID:       userID,
		Username: form.Username,
		Email:    form.Email,
		Password: string(hashedPassword),
	}
	ctx := r.Context()
	if err := app.models.Users.CreateUserAndInvite(ctx, user); err != nil {
		app.errorServerError(w, r, err)
		return
	}
	// todo: send an email to the user for the invite
	// ---------------------------------------------
	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.errorServerError(w, r, err)
		return
	}
}

// activateHandler godoc
//
//	@Summary		Activate a user
//	@Description	Activate a user from their invite token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Activation token"
//	@Success		200		{object}	nil
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/auth/activate/{token} [post]
func (app *application) activateHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	parsedToken, err := uuid.Parse(token)
	if err != nil {
		app.errorBadRequest(w, r, err)
		return
	}

	invite := &models.Invite{InviteToken: parsedToken}
	if err := app.models.Invites.GetInvite(r.Context(), invite); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			app.errorNotFound(w, r, err)
			return
		default:
			app.errorServerError(w, r, err)
			return
		}
	}
	if invite.ExpiresAt.Before(time.Now()) {
		app.errorBadRequest(w, r, errors.New("token expired"))
		return
	}

	// todo: could just create a database call that activates a user based on ID
	ctx := r.Context()
	user, err := app.models.Users.Get(ctx, invite.UserID)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}
	user.IsActivated = true
	// TODO: wrap update and delete in a transaction
	err = app.models.Users.Update(ctx, user)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}
	err = app.models.Invites.Delete(ctx, user.ID)
	if err != nil {
		app.errorServerError(w, r, err)
		return
	}
	if err = app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.errorServerError(w, r, err)
	}
}
