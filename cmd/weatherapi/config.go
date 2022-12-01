package main

import (
	"flag"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/byatesrae/weather/internal/platform/config"
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

	variables := config.New(flag.NewFlagSet(component, flag.ContinueOnError))

	variables.AddIntVar(
		&c.Port,
		"Port",
		8080,
		"The port the service will be listening on.",
		config.AddVarWithEnvName("PORT"),
		config.AddVarWithFlagName("port"))

	variables.AddStringVar(
		&c.OpenweatherEndpointURL,
		"OpenweatherEndpointURL",
		"http://api.openweathermap.org/data/2.5",
		"Endpoint for the Openweather provider API endpoint.",
		config.AddVarWithEnvName("OPENWEATHER_ENDPOINT_URL"),
		config.AddVarWithFlagName("openweather-endpoint-url"))

	variables.AddStringVar(
		&c.OpenweatherAPIKey,
		"OpenweatherAPIKey",
		"",
		"API key for the Openweather provider. See https://openweathermap.org/current.",
		config.AddVarWithEnvName("OPENWEATHER_API_KEY"),
		config.AddVarWithFlagName("openweather-api-key"))

	variables.AddStringVar(
		&c.WeatherstackEndpointURL,
		"WeatherstackEndpointURL",
		"http://api.weatherstack.com",
		"Endpoint for the Weatherstack provider API endpoint.",
		config.AddVarWithEnvName("WEATHERTSTACK_ENDPOINT_URL"),
		config.AddVarWithFlagName("weatherstack-endpoint-url"))

	variables.AddStringVar(
		&c.WeatherstackAccessKey,
		"WeatherstackAccessKey",
		"",
		"Access key for the Weatherstack provider. See https://weatherstack.com/documentation.",
		config.AddVarWithEnvName("WEATHERSTACK_ACCESS_KEY"),
		config.AddVarWithFlagName("weatherstack-access-key"))

	variables.AddDurationVar(
		&c.ResultTimeout,
		"ResultTimeout",
		time.Second*10,
		"Timeout for getting a response from providers.",
		config.AddVarWithEnvName("RESULT_TIMEOUT"),
		config.AddVarWithFlagName("result-timeout"))

	variables.AddDurationVar(
		&c.ResultCacheTTL,
		"ResultCacheTTL",
		time.Second*3,
		"The amount of time a weather result is cached for.",
		config.AddVarWithEnvName("RESULT_CACHE_TTL"), config.AddVarWithFlagName("result-cache-ttl"))

	variables.AddBoolVar(
		&c.ColourizedOutput,
		"ColourizedOutput",
		false,
		"If true, log messages are colourized.",
		config.AddVarWithEnvName("COLOURIZED_OUTPUT"),
		config.AddVarWithFlagName("colourized-output"))

	if err := variables.Parse(os.Args[1:]); err != nil {
		return nil, errors.Wrap(err, "weatherapi: parsing config")
	}

	if c.OpenweatherAPIKey == "" {
		variables.Usage()
		return nil, errors.New("weatherapi: environment variable OPENWEATHER_API_KEY is required")
	}

	if c.WeatherstackAccessKey == "" {
		variables.Usage()
		return nil, errors.New("weatherapi: environment variable WEATHERSTACK_ACCESS_KEY is required")
	}

	return &c, nil
}
