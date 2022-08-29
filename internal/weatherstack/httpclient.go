package weatherstack

//go:generate moq -out httpclient_moq_test.go . HTTPClient

import "net/http"

// HTTPClient is an HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
