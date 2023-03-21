package openweather

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"

	"github.com/byatesrae/weather/internal/platform/nooplogr"
)

// NewOptions are the options for the [New] function.
type NewOptions struct {
	client               HTTPClient
	getLoggerFromContext func(ctx context.Context) logr.Logger
}

// NewWithHTTPClient sets the HTTPClient in the [New] function.
func NewWithHTTPClient(client HTTPClient) func(*NewOptions) {
	return func(o *NewOptions) {
		o.client = client
	}
}

// WithGetLoggerFromContext sets a function used to retrieve a [logr.Logger] from
// the context.
func WithGetLoggerFromContext(getLoggerFromContext func(ctx context.Context) logr.Logger) func(o *NewOptions) {
	return func(o *NewOptions) {
		o.getLoggerFromContext = getLoggerFromContext
	}
}

// Client is used to interact with the Openweather API.
type Client struct {
	client               HTTPClient
	endpointURL          string
	apiKey               string
	getLoggerFromContext func(ctx context.Context) logr.Logger
}

// New creates a new [Client].
func New(endpointURL, apiKey string, optionOverrides ...func(*NewOptions)) *Client {
	noopLogger := nooplogr.New()

	options := NewOptions{
		client: &http.Client{Timeout: time.Second * 5},
		getLoggerFromContext: func(ctx context.Context) logr.Logger {
			return noopLogger
		},
	}

	for _, optionOverride := range optionOverrides {
		optionOverride(&options)
	}

	return &Client{
		client:               options.client,
		apiKey:               apiKey,
		endpointURL:          endpointURL,
		getLoggerFromContext: options.getLoggerFromContext,
	}
}
