package openweather

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceWeatherByCityName(t *testing.T) {
	t.Parallel()

	dummyResult := WeatherSuccess{
		Main: WeatherMain{
			Temperature: 123,
		},
		Wind: WeatherWind{
			WindSpeed: 456,
		},
	}

	for _, tc := range []struct {
		name         string
		withClient   *Client
		giveContext  context.Context
		giveCityName string
		expected     *WeatherSuccess
		expectedErr  string
	}{
		{
			name: "success",
			withClient: New(
				"",
				"",
				NewWithHTTPClient(&HTTPClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						payload, err := json.Marshal(dummyResult)
						assert.NoError(t, err)

						r := io.NopCloser(bytes.NewReader(payload))

						return &http.Response{StatusCode: http.StatusOK, Body: r}, nil
					},
				}),
			),
			giveContext:  context.Background(),
			giveCityName: "Sydney",
			expected:     &dummyResult,
		},
		{
			name: "unexpected_response_type",
			withClient: New(
				"",
				"",
				NewWithHTTPClient(&HTTPClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						payload, err := json.Marshal("ABCDEFG")
						assert.NoError(t, err)

						r := io.NopCloser(bytes.NewReader(payload))

						return &http.Response{StatusCode: http.StatusOK, Body: r}, nil
					},
				}),
			),
			giveContext:  context.Background(),
			giveCityName: "Sydney",
			expected:     nil,
			expectedErr:  "openweather: decode body: json: cannot unmarshal string into Go value of type openweather.WeatherSuccess",
		},
		{
			name: "http_client_error",
			withClient: New(
				"",
				"",
				NewWithHTTPClient(&HTTPClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("intentional test error")
					},
				}),
			),
			giveContext:  context.Background(),
			giveCityName: "Sydney",
			expected:     nil,
			expectedErr:  "openweather: execute request: intentional test error",
		},
		{
			name: "response_code_500",
			withClient: New(
				"",
				"",
				NewWithHTTPClient(&HTTPClientMock{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusInternalServerError}, nil
					},
				}),
			),
			giveContext:  context.Background(),
			giveCityName: "Sydney",
			expected:     nil,
			expectedErr:  "openweather: unexpected response status code 500",
		},
		{
			name:         "missing_city_name",
			withClient:   New("", "", NewWithHTTPClient(&HTTPClientMock{})),
			giveContext:  context.Background(),
			giveCityName: "",
			expected:     nil,
			expectedErr:  "openweather: cityname is required",
		},
		{
			name:         "nil_context",
			withClient:   New("", "", NewWithHTTPClient(&HTTPClientMock{})),
			giveContext:  nil,
			giveCityName: "Sydney",
			expected:     nil,
			expectedErr:  "openweather: create request: net/http: nil Context",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.giveContext
			if tc.giveContext != nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(tc.giveContext)
				t.Cleanup(cancel)
			}

			actual, err := tc.withClient.WeatherByCityName(ctx, tc.giveCityName)

			assert.Equal(t, tc.expected, actual)

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("ctx_cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		client := New("", "", NewWithHTTPClient(&HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()

				return nil, req.Context().Err()
			},
		}))

		actualResult, actualErr := client.WeatherByCityName(ctx, "ABC")
		assert.Nil(t, actualResult)
		assert.ErrorIs(t, actualErr, ctx.Err())
	})
}
