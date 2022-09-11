package main

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeather(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	for _, tc := range []struct {
		name                    string
		withOpenweatherHandler  http.HandlerFunc
		withWeatherstackHandler http.HandlerFunc
		give                    *http.Request
		expectedStatusCode      int
		expectedBody            string
	}{
		{
			name: "success_all",
			withOpenweatherHandler: stubHandler(t, http.StatusOK, []byte(`
			{
				"main": {
					"temp": 4
				},
				"wind": {
					"speed": 2
				}
			}`)),
			withWeatherstackHandler: stubHandler(t, http.StatusOK, []byte(`
			{
				"current": {
					"temperature": 7,
					"wind_speed": 3
				}
			}`)),
			give:               weatherRequest(context.Background(), t, serverURL, "Sydney"),
			expectedStatusCode: http.StatusOK,
			expectedBody:       "{\"wind_speed\":7.2,\"temperature_degrees\":4}\n",
		},
		{
			name: "success_openweather",
			withOpenweatherHandler: stubHandler(t, http.StatusOK, []byte(`
			{
				"main": {
					"temp": 10
				},
				"wind": {
					"speed": 5
				}
			}`)),
			withWeatherstackHandler: stubHandler(t, http.StatusServiceUnavailable, nil),
			give:                    weatherRequest(context.Background(), t, serverURL, "Sydney"),
			expectedStatusCode:      http.StatusOK,
			expectedBody:            "{\"wind_speed\":18,\"temperature_degrees\":10}\n",
		},
		{
			name:                   "success_weatherstack",
			withOpenweatherHandler: stubHandler(t, http.StatusServiceUnavailable, nil),
			withWeatherstackHandler: stubHandler(t, http.StatusOK, []byte(`
			{
				"current": {
					"temperature": 12,
					"wind_speed": 6
				}
			}`)),
			give:               weatherRequest(context.Background(), t, serverURL, "Sydney"),
			expectedStatusCode: http.StatusOK,
			expectedBody:       "{\"wind_speed\":6,\"temperature_degrees\":12}\n",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			// Setup
			requestID := newRequestID(t)

			registerWeatherstackStub(t, requestID, tc.withWeatherstackHandler)
			registerOpenweatherStub(t, requestID, tc.withOpenweatherHandler)

			tc.give.Header.Add("X-Correlation-Id", requestID)

			// Do
			res, err := http.DefaultClient.Do(tc.give)
			require.NoError(t, err, "request error")

			t.Cleanup(func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("failed to close body: %s", err)
				}
			})

			resultExpiry, err := time.Parse(http.TimeFormat, res.Header.Get("Expires"))
			require.NoError(t, err, "parse expiry error")

			// Assert
			actualBody, err := io.ReadAll(res.Body)
			require.NoError(t, err, "read body error")

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode, "response status code")
			assert.Equal(t, tc.expectedBody, string(actualBody), "response body")

			// Wait for cache, with +1 second as cache expiry header resolution
			// is in seconds.
			timeUntilExpiry := time.Until(resultExpiry) + time.Second
			<-time.After(timeUntilExpiry)
		})
	}
}
