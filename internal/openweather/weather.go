package openweather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// WeatherSuccess is a successful response from the Openweather API "Weather" endpoint.
type WeatherSuccess struct {
	Main WeatherMain `json:"main"`
	Wind WeatherWind `json:"wind"`
}

// WeatherMain is part of a successful response from the Openweather API "Weather" endpoint.
type WeatherMain struct {
	Temperature float64 `json:"temp"` // The location temperature in degrees celsius.
}

// WeatherWind is part of a successful response from the Openweather API "Weather" endpoint.
type WeatherWind struct {
	WindSpeed float64 `json:"speed"` // The location windspeed in m/s.
}

// WeatherByCityName returns a summary of the weather for a city.
func (c *Client) WeatherByCityName(ctx context.Context, cityName string) (*WeatherSuccess, error) {
	logger := c.getLoggerFromContext(ctx)

	if cityName == "" {
		return nil, errors.New("openweather: cityname is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/weather", c.endpointURL), http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "openweather: create request")
	}

	q := req.URL.Query()
	q.Add("appid", c.apiKey)
	q.Add("q", cityName)
	q.Add("units", "metric")
	req.URL.RawQuery = q.Encode()

	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "openweather: execute request")
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("openweather: unexpected response status code %v", res.StatusCode)
	}

	var apiResponse WeatherSuccess
	if res.Body != nil {
		defer func() {
			err := res.Body.Close()
			if err != nil {
				logger.Error(err, "Error closing response body.")
			}
		}()

		if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
			return nil, errors.Wrap(err, "openweather: decode body")
		}
	}

	return &apiResponse, nil
}
