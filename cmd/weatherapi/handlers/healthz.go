package handlers

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/byatesrae/weather/internal/platform/nooplogr"
)

// NewHealthzHandler creates a new handler that can be used to check the health status of the application.
func NewHealthzHandler(getLoggerFromContext func(context.Context) logr.Logger) http.HandlerFunc {
	nooplogger := nooplogr.New()

	return func(rw http.ResponseWriter, req *http.Request) {
		logger := nooplogger
		if getLoggerFromContext != nil {
			logger = getLoggerFromContext(req.Context())
		}

		if _, err := rw.Write([]byte("ok")); err != nil {
			logger.Error(err, "Failed to write healthz response body.")

			http.Error(rw, "", http.StatusInternalServerError)
		}
	}
}
