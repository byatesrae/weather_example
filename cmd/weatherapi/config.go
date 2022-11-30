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
	// Port the service will be listening on.
	Port int `env:"PORT"`

	// Endpoint for the Openweather provider API endpoint.
	OpenweatherEndpointURL string `env:"OPENWEATHER_ENDPOINT_URL,default=http://api.openweathermap.org/data/2.5"`

	// API key for the Openweather provider. See https://weatherstack.com/documentation.
	OpenweatherAPIKey string `env:"OPENWEATHER_API_KEY"`

	// Endpoint for the Weatherstack provider API endpoint.
	WeatherstackEndpointURL string `env:"WEATHERSTACK_ENDPOINT_URL,default=http://api.weatherstack.com"`

	// Access key for the Weatherstack provider. See https://weatherstack.com/documentation.
	WeatherstackAccessKey string `env:"WEATHERSTACK_ACCESS_KEY"`

	// Timeout for getting a response from providers.
	ResultTimeout time.Duration `env:"RESULT_TIMEOUT,default=10s"`

	// The amount of time a weather result is cached for.
	ResultCacheTTL time.Duration `env:",default=3s"`

	// If true, log messages are colourized.
	ColourizedOutput bool
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
	var c appConfig

	variables := config.New(flag.NewFlagSet(component, flag.ContinueOnError))
	variables.AddIntVar(
		&c.Port,
		"Port",
		8080,
		"The port the service will be listening on.",
		config.AddVarWithEnvName("PORT"),
		config.AddVarWithFlagName("port"),
	)

	variables.AddStringVar(
		&c.OpenweatherEndpointURL,
		"OpenweatherEndpointURL",
		"http://api.openweathermap.org/data/2.5",
		"Endpoint for the Openweather provider API endpoint.",
		config.AddVarWithEnvName("OPENWEATHER_ENDPOINT_URL"),
		config.AddVarWithFlagName("openweather-endpoint-url"),
	)

	variables.AddStringVar(
		&c.OpenweatherAPIKey,
		"OpenweatherAPIKey",
		"",
		"API key for the Openweather provider. See https://weatherstack.com/documentation.",
		config.AddVarWithEnvName("OPENWEATHER_API_KEY"),
		config.AddVarWithFlagName("openweather-api-key"),
	)
	variables.AddStringVar(&c.WeatherstackEndpointURL, "WeatherstackEndpointURL", "http://api.weatherstack.com", "Endpoint for the Weatherstack provider API endpoint.", config.AddVarWithEnvName("WEATHERTSTACK_ENDPOINT_URL"), config.AddVarWithFlagName("weathertstack-endpoint-url"))
	variables.AddStringVar(&c.WeatherstackAccessKey, "WeatherstackAccessKey", "", "Access key for the Weatherstack provider. See https://weatherstack.com/documentation.", config.AddVarWithEnvName("WEATHERSTACK_ACCESS_KEY"), config.AddVarWithFlagName("weathertstack-access-key"))
	variables.AddDurationVar(&c.ResultTimeout, "ResultTimeout", time.Second*10, "Timeout for getting a response from providers.", config.AddVarWithEnvName("RESULT_TIMEOUT"), config.AddVarWithFlagName("result-timeout"))
	variables.AddDurationVar(&c.ResultCacheTTL, "ResultCacheTTL", time.Second*3, "The amount of time a weather result is cached for.", config.AddVarWithEnvName("RESULT_CACHE_TTL"), config.AddVarWithFlagName("result-cache-ttl"))
	variables.AddBoolVar(&c.ColourizedOutput, "ColourizedOutput", false, "If true, log messages are colourized.", config.AddVarWithEnvName("COLOURIZED_OUTPUT"), config.AddVarWithFlagName("colourized-output"))

	if err := variables.Parse(os.Args[1:]); err != nil {
		return nil, errors.Wrap(err, "weatherapi: parsing config")
	}

	// if err := envconfig.Process(ctx, &config); err != nil {
	// 	return nil, errors.Wrap(err, "weatherapi: loading config")
	// }

	// fs := flag.NewFlagSet(component, flag.ContinueOnError)
	// fs.IntVar(&config.Port, "port", 8080, "The port the service will be listening on.")
	// fs.StringVar(&config.OpenweatherEndpointURL, "openweather-endpoint-url", "http://api.openweathermap.org/data/2.5", "Endpoint for the Openweather provider API endpoint.")
	// fs.StringVar(&config.OpenweatherAPIKey, "openweather-api-key", "", "API key for the Openweather provider. See https://weatherstack.com/documentation.")
	// fs.StringVar(&config.WeatherstackEndpointURL, "weathertstack-endpoint-url", "http://api.weatherstack.com", "Endpoint for the Weatherstack provider API endpoint.")
	// fs.StringVar(&config.WeatherstackAccessKey, "weathertstack-access-key", "", "Access key for the Weatherstack provider. See https://weatherstack.com/documentation.")
	// fs.DurationVar(&config.ResultTimeout, "result-timeout", time.Second*10, "Timeout for getting a response from providers.")
	// fs.DurationVar(&config.ResultCacheTTL, "result-cache-ttl", time.Second*3, "The amount of time a weather result is cached for.")

	// if err := fs.Parse(os.Args[1:]); err != nil {
	// 	return nil, errors.Wrap(err, "weatherapi: parsing config")
	// }

	if c.OpenweatherAPIKey == "" {
		return nil, errors.New("weatherapi: environment variable OPENWEATHER_API_KEY is required")
	}

	if c.WeatherstackAccessKey == "" {
		return nil, errors.New("weatherapi: environment variable WEATHERSTACK_ACCESS_KEY is required")
	}

	return &c, nil
}
