package openweather

import "net/http"

// HTTPClient is an HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
