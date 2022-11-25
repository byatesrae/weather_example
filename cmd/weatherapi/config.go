package main

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
)

// appConfig is all of the application configuration.
type appConfig struct {
	// Port the service will be listening on.
	Port int `env:"PORT,default=8080"`

	// Endpoint for the Openweather provider API endpoint.
	OpenweatherEndpointURL string `env:"OPENWEATHER_ENDPOINT_URL,default=http://api.openweathermap.org/data/2.5"`

	// API key for the Openweather provider. See https://weatherstack.com/documentation.
	OpenweatherAPIKey string `env:"OPENWEATHER_API_KEY,required"`

	// Endpoint for the Weatherstack provider API endpoint.
	WeatherstackEndpointURL string `env:"WEATHERTSTACK_ENDPOINT_URL,default=http://api.weatherstack.com"`

	// Access key for the Weatherstack provider. See https://weatherstack.com/documentation.
	WeatherstackAccessKey string `env:"WEATHERTSTACK_ACCESS_KEY,required"`

	// Timeout for getting a response from providers.
	ResultTimeout time.Duration `env:"RESULT_TIMEOUT,default=10s"`

	// The amount of time a weather result is cached for.
	ResultCacheTTL time.Duration `env:"RESULT_CACHE_TTL,default=3s"`
}

func (c *appConfig) masked() *appConfig {
	masked := *c

	if masked.OpenweatherAPIKey != "" {
		masked.OpenweatherAPIKey = "*****"
	}

	if masked.WeatherstackAccessKey != "" {
		masked.WeatherstackAccessKey = "*****"
	}

	return &masked
}

// loadConfig loads the application configuration from environment variables.
func loadConfig(ctx context.Context) (*appConfig, error) {
	var config appConfig
	if err := envconfig.Process(ctx, &config); err != nil {
		return nil, errors.Wrap(err, "weatherapi: loading config")
	}

	if config.OpenweatherAPIKey == "" {
		return nil, errors.New("weatherapi: environment variable OPENWEATHER_API_KEY is required")
	}

	if config.WeatherstackAccessKey == "" {
		return nil, errors.New("weatherapi: environment variable WEATHERTSTACK_ACCESS_KEY is required")
	}

	return &config, nil
}
