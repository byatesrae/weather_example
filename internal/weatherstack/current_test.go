package weatherstack

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

func TestServiceCurrentByCityName(t *testing.T) {
	t.Parallel()

	dummyResult := CurrentSuccess{
		Current: CurrentWeather{
			WindSpeed:   123,
			Temperature: 456,
		},
	}

	for _, tc := range []struct {
		name         string
		withClient   *Client
		giveContext  context.Context
		giveCityName string
		expected     *CurrentSuccess
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
			expectedErr:  "weatherstack: decode body: json: cannot unmarshal string into Go value of type weatherstack.CurrentSuccess",
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
			expectedErr:  "weatherstack: execute request: intentional test error",
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
			expectedErr:  "weatherstack: unexpected response status code 500",
		},
		{
			name:         "missing_city_name",
			withClient:   New("", "", NewWithHTTPClient(&HTTPClientMock{})),
			giveContext:  context.Background(),
			giveCityName: "",
			expected:     nil,
			expectedErr:  "weatherstack: cityname is required",
		},
		{
			name:         "nil_context",
			withClient:   New("", "", NewWithHTTPClient(&HTTPClientMock{})),
			giveContext:  nil,
			giveCityName: "Sydney",
			expected:     nil,
			expectedErr:  "weatherstack: create request: net/http: nil Context",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.giveContext
			if tc.giveContext != nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(tc.giveContext)
				t.Cleanup(cancel)
			}

			actual, err := tc.withClient.CurrentByCityName(ctx, tc.giveCityName)

			assert.Equal(t, tc.expected, actual)

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("ctx_cancel", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		client := New("", "", NewWithHTTPClient(&HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()

				return nil, req.Context().Err()
			},
		}))

		actualResult, actualErr := client.CurrentByCityName(ctx, "ABC")
		assert.Nil(t, actualResult)
		assert.ErrorIs(t, actualErr, ctx.Err())
	})
}
