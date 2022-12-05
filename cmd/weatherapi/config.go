package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/byatesrae/weather/internal/platform/startupconfig"
)

// appConfig is all of the application configuration.
type appConfig struct {
	Port                    int           // Port the service will be listening on.
	OpenweatherEndpointURL  string        // Endpoint for the Openweather provider API endpoint.
	OpenweatherAPIKey       string        // API key for the Openweather provider. See https://weatherstack.com/documentation.
	WeatherstackEndpointURL string        // Endpoint for the Weatherstack provider API endpoint.
	WeatherstackAccessKey   string        // Access key for the Weatherstack provider. See https://weatherstack.com/documentation.
	ResultTimeout           time.Duration // Timeout for getting a response from providers.
	ResultCacheTTL          time.Duration // The amount of time a weather result is cached for.
	ColourizedOutput        bool          // If true, log messages are colourized.
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
func loadConfig() (*appConfig, error) {
	c := appConfig{}

	fs := flag.NewFlagSet(component, flag.ContinueOnError)

	p := startupconfig.Parser{Fs: fs}
	fs.Usage = p.Usage

	fs.IntVar(&c.Port, "port", 8080, "The port the service will be listening on.")
	fs.StringVar(&c.OpenweatherEndpointURL, "openweather-endpoint-url", "http://api.openweathermap.org/data/2.5", "Endpoint for the Openweather provider API endpoint.")
	fs.StringVar(&c.OpenweatherAPIKey, "openweather-api-key", "", "Required. API key for the Openweather provider. See https://openweathermap.org/current.")
	fs.StringVar(&c.WeatherstackEndpointURL, "weatherstack-endpoint-url", "http://api.weatherstack.com", "Endpoint for the Weatherstack provider API endpoint.")
	fs.StringVar(&c.WeatherstackAccessKey, "weatherstack-access-key", "", "Required. Access key for the Weatherstack provider. See https://weatherstack.com/documentation.")
	fs.DurationVar(&c.ResultTimeout, "result-timeout", time.Second*10, "Timeout for getting a response from providers.")
	fs.DurationVar(&c.ResultCacheTTL, "result-cache-ttl", time.Second*3, "The amount of time a weather result is cached for.")
	fs.BoolVar(&c.ColourizedOutput, "colourized-output", false, "If true, log messages are colourized.")

	if err := p.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}

		return nil, errors.Wrap(err, "weatherapi: parsing config")
	}

	if c.OpenweatherAPIKey == "" {
		return nil, fmt.Errorf("weatherapi: validate configuration: %w", p.FlagError("openweather-api-key", fmt.Errorf("value is required")))
	}

	if c.WeatherstackAccessKey == "" {
		return nil, fmt.Errorf("weatherapi: validate configuration: %w", p.FlagError("weatherstack-access-key", fmt.Errorf("value is required")))
	}

	return &c, nil
}
