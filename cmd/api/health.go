package main

import "net/http"

type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Env     string `json:"env" example:"development"`
	Version string `json:"version" example:"0.0.1"`
}

// health godoc
//
//	@Summary		Health check
//	@Description	Returns service health information
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health [get]
func (app *application) health(w http.ResponseWriter, r *http.Request) {
	data := HealthResponse{
		Status:  "ok",
		Env:     app.config.env,
		Version: version,
	}
	err := writeJSON(w, http.StatusOK, data)
	if err != nil {
		app.errorServerError(w, r, err)
	}
}
