package openweather

import (
	"log"
	"net/http"
	"time"
)

// NewOptions are the options for the New function.
type NewOptions struct {
	client HTTPClient
}

// NewWithHTTPClient sets the HTTPClient in the New function.
func NewWithHTTPClient(client HTTPClient) func(*NewOptions) {
	return func(o *NewOptions) {
		o.client = client
	}
}

// Client is used to interact with the Openweather API.
type Client struct {
	client      HTTPClient
	endpointURL string
	apiKey      string
	logger      *log.Logger
}

// New creates a new client.
func New(endpointURL, apiKey string, optionOverrides ...func(*NewOptions)) *Client {
	options := NewOptions{
		client: &http.Client{Timeout: time.Second * 5},
	}
	for _, optionOverride := range optionOverrides {
		optionOverride(&options)
	}

	return &Client{
		client:      options.client,
		apiKey:      apiKey,
		endpointURL: endpointURL,
		logger:      log.Default(),
	}
}
