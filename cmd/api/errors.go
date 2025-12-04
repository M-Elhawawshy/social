package main

import (
	"fmt"
	"net/http"
)

func (app *application) errorBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("%s: %s: %s error: %s\n", http.StatusText(http.StatusBadRequest), r.Method, r.URL.Path, err)
	_ = WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("%s", err))
}

func (app *application) errorServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("%s: %s: %s error: %s\n", http.StatusText(http.StatusInternalServerError), r.Method, r.URL.Path, err)
	_ = WriteJSONError(w, http.StatusInternalServerError, fmt.Sprintf("%s", http.StatusText(http.StatusInternalServerError)))
}

func (app *application) errorNotFound(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("%s: %s: %s error: %s\n", http.StatusText(http.StatusNotFound), r.Method, r.URL.Path, err)
	_ = WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("%s", http.StatusText(http.StatusNotFound)))
}
