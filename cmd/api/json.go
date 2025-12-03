package main

import (
	"encoding/json"
	"net/http"
	"social/internal/models"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

// ErrorResponse represents a standard error message returned by the API.
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Message describing the error
	// example: invalid request payload
	Error string `json:"error"`
}

// DataResponsePost wraps a Post in the standard data envelope.
// swagger:model DataResponsePost
type DataResponsePost struct {
	Data models.Post `json:"data"`
}

// DataResponseFeed wraps a list of feed posts in the standard data envelope.
// swagger:model DataResponseFeed
type DataResponseFeed struct {
	Data []models.FeedPost `json:"data"`
}

// DataResponseComment wraps a Comment in the standard data envelope.
// swagger:model DataResponseComment
type DataResponseComment struct {
	Data models.Comment `json:"data"`
}

// DataResponseUser wraps a User in the standard data envelope.
// swagger:model DataResponseUser
type DataResponseUser struct {
	Data models.User `json:"data"`
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

func WriteJSONError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, ErrorResponse{Error: message})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type dataWrapper struct {
		Data any `json:"data"`
	}
	return writeJSON(w, status, dataWrapper{Data: data})
}
