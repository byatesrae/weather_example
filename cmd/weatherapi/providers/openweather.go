package providers

import (
	"context"
	"fmt"

	"github.com/byatesrae/weather"
	"github.com/byatesrae/weather/internal/openweather"
	"github.com/byatesrae/weather/internal/providerquery"
)

// OpenWeatherProvider wraps an [openweather.Client] to satisfy the [providerquery.Provider] interface.
type OpenWeatherProvider struct {
	client *openweather.Client
}

var _ providerquery.Provider = (*OpenWeatherProvider)(nil)

// NewOpenWeatherProvider creates a new [OpenWeatherProvider].
func NewOpenWeatherProvider(w *openweather.Client) *OpenWeatherProvider {
	return &OpenWeatherProvider{client: w}
}

// ProviderName is the unique name for this provider.
func (p *OpenWeatherProvider) ProviderName() string {
	return "Openweather"
}

// GetWeatherSummary gets a [weather.Summary] for a city.
func (p *OpenWeatherProvider) GetWeatherSummary(ctx context.Context, cityName string) (*weather.Summary, error) {
	res, err := p.client.WeatherByCityName(ctx, cityName)
	if err != nil {
		return nil, fmt.Errorf("get weather by city name: %w", err)
	}

	return &weather.Summary{
		Temperature: res.Main.Temperature,
		WindSpeed:   (res.Wind.WindSpeed * 60 * 60) / 1000, // translate between m/s to km/h
	}, nil
}
