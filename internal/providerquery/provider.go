package providerquery

import (
	"context"

	"github.com/byatesrae/weather"
)

// Provider can be queried for weather summaries.
type Provider interface {
	// ProviderName is a unique name for the provider.
	ProviderName() string

	// GetWeatherSummary gets a weather summary for a city.
	GetWeatherSummary(ctx context.Context, cityName string) (*weather.Summary, error)
}
