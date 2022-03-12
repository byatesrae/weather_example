package handlers

import (
	"log"
	"net/http"
)

// NewHealthzHandler creates a new handler that can be used to check the health status of the application.
func NewHealthzHandler() http.HandlerFunc {
	logger := log.Default()

	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			logger.Printf("[ERR] Health: %s", err)
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
}
