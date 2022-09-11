package main

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// requestCount is used for generating a unique request ID.
var requestCount uint64

// newRequestID generates a unique request ID.
func newRequestID(t *testing.T) string {
	t.Helper()

	myid := atomic.AddUint64(&requestCount, 1)

	requestID := fmt.Sprintf("%s_%06d", t.Name(), myid)

	t.Logf("RequestID: %s", requestID)

	return requestID
}

// registerWeatherstackStub calls weatherstackStubServerRegister.Register(),
// checks the error and handles the cleanup.
func registerWeatherstackStub(t *testing.T, requestID string, h http.Handler) {
	t.Helper()

	cleanup, err := weatherstackStubServerHandler.Register(requestID, h)
	require.NoError(t, err, "weatherstackStubServerHandler register error")

	t.Cleanup(cleanup)
}

// registerOpenweatherStub calls openweatherStubServerHandler.Register(),
// checks the error and handles the cleanup.
func registerOpenweatherStub(t *testing.T, requestID string, h http.Handler) {
	t.Helper()

	cleanup, err := openweatherStubServerHandler.Register(requestID, h)
	require.NoError(t, err, "openweatherStubServerHandler register error")

	t.Cleanup(cleanup)
}

// weatherRequest creates an http request for the weather endpoint.
func weatherRequest(ctx context.Context, t *testing.T, serverURL, city string) *http.Request {
	t.Helper()

	url := fmt.Sprintf("%s/v1/weather", serverURL)

	if city != "" {
		url = fmt.Sprintf("%s?city=%s", url, city)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	require.NoError(t, err, "create weather request")

	return req
}

// stubHandler creates a new http handler that returns a status code & body.
func stubHandler(t *testing.T, statuscode int, body []byte) http.HandlerFunc {
	t.Helper()

	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(statuscode)

		if len(body) > 0 {
			c, err := rw.Write(body)

			require.Equal(t, c, len(body), "length of body write")
			require.NoError(t, err, "body write error")
		}
	}
}
