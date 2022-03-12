package providerquery

import (
	"context"

	"github.com/pkg/errors"

	"github.com/byatesrae/weather"
)

// mockProvider mocks Provider
type mockProvider struct {
	providerName      func() string
	getWeatherSummary func(ctx context.Context, cityName string) (*weather.Summary, error)
}

var _ Provider = (*mockProvider)(nil)

func (m *mockProvider) ProviderName() string {
	if m.providerName != nil {
		return m.providerName()
	}

	return ""
}

func (m *mockProvider) GetWeatherSummary(ctx context.Context, cityName string) (*weather.Summary, error) {
	if m.getWeatherSummary != nil {
		return m.getWeatherSummary(ctx, cityName)
	}

	return nil, errors.New("not implemented")
}
