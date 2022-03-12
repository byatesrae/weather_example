package openweather

import (
	"net/http"

	"github.com/pkg/errors"
)

// mockHTTPClient mocks HTTPClient
type mockHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)
}

var _ HTTPClient = (*mockHTTPClient)(nil)

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.do != nil {
		return m.do(req)
	}

	return nil, errors.New("not implemented")
}
