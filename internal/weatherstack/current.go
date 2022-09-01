package weatherstack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// CurrentSuccess is a successful response from the Weatherstack API "Current" endpoint.
type CurrentSuccess struct {
	Current CurrentWeather `json:"current"`
}

// CurrentWeather is part of a successful response from the Weatherstack API "Current" endpoint.
type CurrentWeather struct {
	Temperature int `json:"temperature"` // The location temperature in degrees celsius.
	WindSpeed   int `json:"wind_speed"`  // The location windspeed in km/h.
}

// CurrentByCityName returns a summary of the weather for a city.
func (c *Client) CurrentByCityName(ctx context.Context, cityName string) (*CurrentSuccess, error) {
	if cityName == "" {
		return nil, errors.New("weatherstack: cityname is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/current", c.endpointURL), http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "weatherstack: create request")
	}

	q := req.URL.Query()
	q.Add("access_key", c.accessKey)
	q.Add("query", cityName)
	q.Add("units", "m")
	req.URL.RawQuery = q.Encode()

	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "weatherstack: execute request")
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("weatherstack: unexpected response status code %v", res.StatusCode)
	}

	var apiResponse CurrentSuccess
	if res.Body != nil {
		defer func() {
			err := res.Body.Close()
			if err != nil {
				c.logger.Printf("weatherstack: err when closing response body: %s\n", err)
			}
		}()

		if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
			return nil, errors.Wrap(err, "weatherstack: decode body")
		}
	}

	return &apiResponse, nil
}
