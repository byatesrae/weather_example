package providers

import (
	"context"
	"fmt"

	"github.com/byatesrae/weather"
	"github.com/byatesrae/weather/internal/providerquery"
	"github.com/byatesrae/weather/internal/weatherstack"
)

// WeatherStackProvider wraps a WeatherStack client to satisfy the providerquery.Provider interface.
type WeatherStackProvider struct {
	client *weatherstack.Client
}

var _ providerquery.Provider = (*WeatherStackProvider)(nil)

// NewWeatherStackProvider creates a new WeatherStackProvider.
func NewWeatherStackProvider(w *weatherstack.Client) *WeatherStackProvider {
	return &WeatherStackProvider{client: w}
}

// ProviderName is the unique name for this provider.
func (p *WeatherStackProvider) ProviderName() string {
	return "Weatherstack"
}

// GetWeatherSummary gets a weather summary for a city.
func (p *WeatherStackProvider) GetWeatherSummary(ctx context.Context, cityName string) (*weather.Summary, error) {
	res, err := p.client.CurrentByCityName(ctx, cityName)
	if err != nil {
		return nil, fmt.Errorf("current by city name: %w", err)
	}

	return &weather.Summary{
		Temperature: float64(res.Current.Temperature),
		WindSpeed:   float64(res.Current.WindSpeed),
	}, nil
}
