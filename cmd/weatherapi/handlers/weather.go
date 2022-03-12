package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/byatesrae/weather/internal/providerquery"
)

const supportedCity = "Sydney"

// ErrorResponse is returned from the API in the event of an error.
type ErrorResponse struct {
	Message string `json:"msg"`
}

// WeatherService is used to query weather for a city.
type WeatherService interface {
	ReadWeatherResult(ctx context.Context, city string) (*providerquery.WeatherResult, error)
}

// NewWeatherHandler creates a new handler that can be used to query a weather summary for a city location.
func NewWeatherHandler(weatherService WeatherService, loadResultTimeout time.Duration) http.HandlerFunc {
	logger := log.Default()

	return func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		if city == "" {
			errorResponse(logger, w, "Missing parameter \"city\".", http.StatusBadRequest)
			return
		}

		if city != supportedCity {
			errorResponse(
				logger,
				w,
				fmt.Sprintf("City %q is not supported. Only %q is currently supported.", city, supportedCity),
				http.StatusBadRequest,
			)

			return
		}

		readWeatherCtx, readWeatherCancel := context.WithTimeout(r.Context(), loadResultTimeout)
		defer readWeatherCancel()

		result, err := weatherService.ReadWeatherResult(readWeatherCtx, city)
		if err != nil {
			errorResponse(
				logger,
				w,
				"Woops, something went wrong.",
				http.StatusInternalServerError,
			)
		}

		if result != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "public")
			w.Header().Set("Last-modified", result.CreatedAt.Format(http.TimeFormat))
			w.Header().Set("Expires", result.Expiry.Format(http.TimeFormat))

			if err := json.NewEncoder(w).Encode(result.Weather); err != nil {
				logger.Printf("[ERR] Failed to encode body: %s\n", err)

				http.Error(w, "", http.StatusInternalServerError)
			}
		}
	}
}

func errorResponse(logger *log.Logger, w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(&ErrorResponse{Message: message}); err != nil {
		logger.Printf("[ERR] Failed to encode error response body: %s\n", err)

		http.Error(w, message, code)
	}
}
