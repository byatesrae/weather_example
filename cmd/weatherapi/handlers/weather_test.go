package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/byatesrae/weather"
	"github.com/byatesrae/weather/internal/providerquery"
)

func TestWeatherHandler(t *testing.T) {
	t.Parallel()

	now := time.Date(2020, time.November, 11, 10, 10, 10, 0, time.UTC)
	goodServiceResult := &providerquery.WeatherResult{
		Weather:   &weather.Summary{Temperature: 123.456},
		CreatedAt: now,
		Expiry:    now.Add(time.Second * 5),
	}
	goodService := &WeatherServiceMock{
		ReadWeatherResultFunc: func(ctx context.Context, city string) (*providerquery.WeatherResult, error) {
			return goodServiceResult, nil
		},
	}
	errService := &WeatherServiceMock{
		ReadWeatherResultFunc: func(ctx context.Context, city string) (*providerquery.WeatherResult, error) {
			return nil, errors.New("intentional test error")
		},
	}
	hangService := &WeatherServiceMock{
		ReadWeatherResultFunc: func(ctx context.Context, city string) (*providerquery.WeatherResult, error) {
			<-ctx.Done()

			return nil, ctx.Err()
		},
	}

	goodRequest := httptest.NewRequest("GET", "/weather?city=Sydney", nil)

	for _, tc := range []struct {
		name         string
		withHandler  http.HandlerFunc
		giveRequest  *http.Request
		expectedCode int
		expectedBody []byte
	}{
		{
			name:         "success",
			withHandler:  NewWeatherHandler(goodService, time.Millisecond*100),
			giveRequest:  goodRequest,
			expectedCode: http.StatusOK,
			expectedBody: []byte("{\"wind_speed\":0,\"temperature_degrees\":123.456}\n"),
		},
		{
			name:         "city_empty",
			withHandler:  NewWeatherHandler(goodService, time.Millisecond*100),
			giveRequest:  httptest.NewRequest("GET", "/weather", nil),
			expectedCode: http.StatusBadRequest,
			expectedBody: []byte("{\"msg\":\"Missing parameter \\\"city\\\".\"}\n"),
		},
		{
			name:         "city_invalid",
			withHandler:  NewWeatherHandler(goodService, time.Millisecond*100),
			giveRequest:  httptest.NewRequest("GET", "/weather?city=abc", nil),
			expectedCode: http.StatusBadRequest,
			expectedBody: []byte("{\"msg\":\"City \\\"abc\\\" is not supported. Only \\\"Sydney\\\" is currently supported.\"}\n"),
		},
		{
			name:         "weather_service_err",
			withHandler:  NewWeatherHandler(errService, time.Millisecond*100),
			giveRequest:  goodRequest,
			expectedCode: http.StatusInternalServerError,
			expectedBody: []byte("{\"msg\":\"Woops, something went wrong.\"}\n"),
		},
		{
			name:         "weather_service_hang",
			withHandler:  NewWeatherHandler(hangService, time.Millisecond*100),
			giveRequest:  goodRequest,
			expectedCode: http.StatusInternalServerError,
			expectedBody: []byte("{\"msg\":\"Woops, something went wrong.\"}\n"),
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()

			tc.withHandler.ServeHTTP(rr, tc.giveRequest)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.Equal(t, string(tc.expectedBody), rr.Body.String())
		})
	}
}
