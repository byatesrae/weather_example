package main

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
)

// config is all of the application configuration.
type config struct {
	Port                    int           `env:"PORT,default=8080"`                                                       // Port the service will be hosted on.
	OpenweatherEndpointURL  string        `env:"OPENWEATHER_ENDPOINT_URL,default=http://api.openweathermap.org/data/2.5"` // Endpoint for the Openweather provider API endpoint.
	OpenweatherAPIKey       string        `env:"OPENWEATHER_API_KEY,required"`                                            // API key for the Openweather provider. See https://weatherstack.com/documentation.
	WeatherstackEndpointURL string        `env:"WEATHERTSTACK_ENDPOINT_URL,default=http://api.weatherstack.com"`          // Endpoint for the Weatherstack provider API endpoint.
	WeatherstackAccessKey   string        `env:"WEATHERTSTACK_ACCESS_KEY,required"`                                       // Access key for the Weatherstack provider. See https://weatherstack.com/documentation.
	ResultTimeout           time.Duration `env:"RESULT_TIMEOUT,default=10s"`                                              // Timeout for getting a response from providers.
}

// loadConfig loads the application configuration from environment variables.
func loadConfig(ctx context.Context) (*config, error) {
	var c config
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, errors.Wrap(err, "weatherapi: loading config")
	}

	if c.OpenweatherAPIKey == "" {
		return nil, errors.New("weatherapi: environment variable OPENWEATHER_API_KEY is required")
	}

	if c.WeatherstackAccessKey == "" {
		return nil, errors.New("weatherapi: environment variable WEATHERTSTACK_ACCESS_KEY is required")
	}

	return &c, nil
}
