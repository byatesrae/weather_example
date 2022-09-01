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

	return func(rw http.ResponseWriter, req *http.Request) {
		city := req.URL.Query().Get("city")
		if city == "" {
			errorResponse(logger, rw, "Missing parameter \"city\".", http.StatusBadRequest)

			return
		}

		if city != supportedCity {
			errorResponse(
				logger,
				rw,
				fmt.Sprintf("City %q is not supported. Only %q is currently supported.", city, supportedCity),
				http.StatusBadRequest,
			)

			return
		}

		readWeatherCtx, readWeatherCancel := context.WithTimeout(req.Context(), loadResultTimeout)
		defer readWeatherCancel()

		result, err := weatherService.ReadWeatherResult(readWeatherCtx, city)
		if err != nil {
			errorResponse(
				logger,
				rw,
				"Woops, something went wrong.",
				http.StatusInternalServerError,
			)
		}

		if result != nil {
			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("Cache-Control", "public")
			rw.Header().Set("Last-modified", result.CreatedAt.Format(http.TimeFormat))
			rw.Header().Set("Expires", result.Expiry.Format(http.TimeFormat))

			if err := json.NewEncoder(rw).Encode(result.Weather); err != nil {
				logger.Printf("[ERR] Failed to encode body: %s\n", err)

				http.Error(rw, "", http.StatusInternalServerError)
			}
		}
	}
}

func errorResponse(logger *log.Logger, rw http.ResponseWriter, message string, code int) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)

	if err := json.NewEncoder(rw).Encode(&ErrorResponse{Message: message}); err != nil {
		logger.Printf("[ERR] Failed to encode error response body: %s\n", err)

		http.Error(rw, message, code)
	}
}
