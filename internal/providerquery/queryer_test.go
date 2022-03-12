package providerquery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/byatesrae/weather"
)

// mockCache mocks Cache
type mockCache struct {
	get func(ctx context.Context, key interface{}) (interface{}, time.Time, error)
	set func(ctx context.Context, key, val interface{}, expiry time.Time) error
}

func (m *mockCache) Get(ctx context.Context, key interface{}) (interface{}, time.Time, error) {
	if m.get != nil {
		return m.get(ctx, key)
	}

	return nil, time.Time{}, nil
}

func (m *mockCache) Set(ctx context.Context, key, val interface{}, expiry time.Time) error {
	if m.set != nil {
		return m.set(ctx, key, val, expiry)
	}

	return nil
}

func TestQueryerReadWeatherResult(t *testing.T) {
	t.Parallel()

	clock := fixedClock{now: time.Date(2020, time.November, 11, 10, 10, 10, 0, time.UTC)}

	goodResult := &WeatherResult{
		Weather:   &weather.Summary{Temperature: 123.456},
		CreatedAt: clock.now,
		Expiry:    clock.now.Add(resultTTL),
	}

	goodProvider := &mockProvider{
		getWeatherSummary: func(ctx context.Context, cityName string) (*weather.Summary, error) {
			return goodResult.Weather, nil
		},
	}

	errProvider := &mockProvider{
		getWeatherSummary: func(ctx context.Context, cityName string) (*weather.Summary, error) {
			return nil, errors.New("intentional test error")
		},
	}

	hangingProvider := &mockProvider{
		getWeatherSummary: func(ctx context.Context, cityName string) (*weather.Summary, error) {
			<-ctx.Done()

			return nil, ctx.Err()
		},
	}

	emptyCache := &mockCache{
		get: func(ctx context.Context, key interface{}) (interface{}, time.Time, error) {
			return nil, time.Time{}, nil
		},
		set: func(ctx context.Context, key, val interface{}, expiry time.Time) error {
			return nil
		},
	}
	cacheWithResult := &mockCache{
		get: func(ctx context.Context, key interface{}) (interface{}, time.Time, error) {
			return resultCacheEntry{result: goodResult.Weather, createdAt: goodResult.CreatedAt}, goodResult.Expiry, nil
		},
		set: func(ctx context.Context, key, val interface{}, expiry time.Time) error {
			return nil
		},
	}
	errCache := &mockCache{
		get: func(ctx context.Context, key interface{}) (interface{}, time.Time, error) {
			return nil, time.Time{}, errors.New("intentional test error")
		},
		set: func(ctx context.Context, key, val interface{}, expiry time.Time) error {
			return errors.New("intentional test error")
		},
	}
	hangCache := &mockCache{
		get: func(ctx context.Context, key interface{}) (interface{}, time.Time, error) {
			<-ctx.Done()

			return nil, time.Time{}, ctx.Err()
		},
		set: func(ctx context.Context, key, val interface{}, expiry time.Time) error {
			<-ctx.Done()

			return ctx.Err()
		},
	}

	for _, tc := range []struct {
		name        string
		withQueryer *Queryer
		giveContext context.Context
		giveCity    string
		expected    *WeatherResult
		expectedErr string
	}{
		{
			name:        "success",
			withQueryer: New([]Provider{goodProvider}, emptyCache, withClock(clock)),
			giveContext: context.Background(),
			giveCity:    "ABC",
			expected:    goodResult,
		},
		{
			name:        "success_cache_error",
			withQueryer: New([]Provider{goodProvider}, errCache, withClock(clock)),
			giveContext: context.Background(),
			giveCity:    "ABC",
			expected:    goodResult,
		},
		{
			name:        "success_cache_hang",
			withQueryer: New([]Provider{goodProvider}, hangCache, withClock(clock)),
			giveContext: context.Background(),
			giveCity:    "ABC",
			expected:    goodResult,
		},
		{
			name:        "success_use_cache_provider_err",
			withQueryer: New([]Provider{errProvider}, cacheWithResult, withClock(clock)),
			giveContext: context.Background(),
			giveCity:    "ABC",
			expected:    goodResult,
		},
		{
			name:        "success_use_cache_provider_hang",
			withQueryer: New([]Provider{hangingProvider}, cacheWithResult, withClock(clock)),
			giveContext: context.Background(),
			giveCity:    "ABC",
			expected:    goodResult,
		},
		{
			name:        "provider_err_empty_cache",
			withQueryer: New([]Provider{errProvider}, emptyCache, withClock(clock)),
			giveContext: context.Background(),
			giveCity:    "ABC",
			expectedErr: "providerqueryer: failed to load a new result and no cached result to fall back on",
		},
		{
			name:        "provider_err_cache_err",
			withQueryer: New([]Provider{errProvider}, errCache, withClock(clock)),
			giveContext: context.Background(),
			giveCity:    "ABC",
			expectedErr: "providerqueryer: failed to load a new result and no cached result to fall back on",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(tc.giveContext)
			t.Cleanup(cancel)

			actual, actualErr := tc.withQueryer.ReadWeatherResult(ctx, tc.giveCity)
			assert.Equal(t, tc.expected, actual)

			if tc.expectedErr != "" {
				assert.EqualError(t, actualErr, tc.expectedErr)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}
