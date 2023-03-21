package providerquery

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"

	"github.com/byatesrae/weather"
	"github.com/byatesrae/weather/internal/platform/nooplogr"
)

const (
	providerLogKey = "provider"
)

// WeatherResult contains a weather summary and timeline data.
type WeatherResult struct {
	Expiry    time.Time
	CreatedAt time.Time
	Weather   *weather.Summary
}

// Queryer will query a list of providers for a weather summary.
type Queryer struct {
	getLoggerFromContext func(ctx context.Context) logr.Logger
	cache                Cache

	// Timeout for querying the cache.
	cacheTimeout time.Duration

	// A slice of providers to query, ordered by query preference.
	providers []Provider

	// Timeout for querying an individual provider.
	providerTimeout time.Duration

	// Regardless of how many times ReadWeatherResult is called, query providers once (to avoid a thundering heard).
	queryAllProvidersForWeatherOnce singleflight.Group

	// TTL applied for cached provider results.
	resultCacheTTL time.Duration
	resultTimeout  time.Duration

	// Timeout for getting a response across all providers.
	clock Clock
}

// NewOptions are options for the New function.
type NewOptions struct {
	clock                Clock
	resultCacheTTL       time.Duration
	getLoggerFromContext func(ctx context.Context) logr.Logger
}

// withClock sets the clock used in the New function.
func withClock(clock Clock) func(o *NewOptions) {
	return func(o *NewOptions) {
		o.clock = clock
	}
}

// WithResultCacheTTL sets the amount of time a result is cached for.
func WithResultCacheTTL(resultCacheTTL time.Duration) func(o *NewOptions) {
	return func(o *NewOptions) {
		o.resultCacheTTL = resultCacheTTL
	}
}

// WithGetLoggerFromContext sets a function used to retrieve a [logr.Logger] from
// the context.
func WithGetLoggerFromContext(getLoggerFromContext func(ctx context.Context) logr.Logger) func(o *NewOptions) {
	return func(o *NewOptions) {
		o.getLoggerFromContext = getLoggerFromContext
	}
}

// New creates a new [Queryer].
func New(providers []Provider, cache Cache, overrides ...func(o *NewOptions)) *Queryer {
	noopLogger := nooplogr.New()

	options := &NewOptions{
		clock:          standardClock{},
		resultCacheTTL: time.Second * 3,
		getLoggerFromContext: func(ctx context.Context) logr.Logger {
			return noopLogger
		},
	}

	for _, override := range overrides {
		override(options)
	}

	return &Queryer{
		getLoggerFromContext: options.getLoggerFromContext,
		cache:                cache,
		cacheTimeout:         time.Second * 2, // These timeouts should all be configurable.
		providers:            providers,
		providerTimeout:      time.Second * 3,
		resultCacheTTL:       options.resultCacheTTL,
		resultTimeout:        time.Second * 10,
		clock:                options.clock,
	}
}

// ReadWeatherResult will query one or more providers for a weather result. The result
// will be cached and sometimes served stale.
func (q *Queryer) ReadWeatherResult(ctx context.Context, city string) (*WeatherResult, error) {
	logger := q.getLoggerFromContext(ctx)

	result := q.getCachedReadWeatherResult(ctx, logger)

	retrievedCachedResult := result != nil

	if !retrievedCachedResult || q.clock.Now().After(result.Expiry) {
		logger.V(1).Info("Querying all providers.")

		queryAllProvidersCtx, queryAllProvidersCancel := context.WithTimeout(ctx, q.resultTimeout)
		defer queryAllProvidersCancel()

		newWeather, err, _ := q.queryAllProvidersForWeatherOnce.Do("queryAllProvidersForWeather", func() (interface{}, error) {
			return q.queryAllProvidersForWeather(queryAllProvidersCtx, logger, city)
		})
		if err != nil {
			logger.Error(err, "Failed to retrieve new weather result.")

			if !retrievedCachedResult {
				return nil, errors.New("providerqueryer: failed to load a new result and no cached result to fall back on")
			}
		} else if newWeather != nil {
			now := q.clock.Now().UTC()
			result = &WeatherResult{
				Weather:   newWeather.(*weather.Summary), // should never panic
				CreatedAt: now,
				Expiry:    now.Add(q.resultCacheTTL),
			}

			go q.cacheWeatherResult(ctx, logger, result)
		}
	}

	return result, nil
}

func (q *Queryer) getCachedReadWeatherResult(ctx context.Context, logger logr.Logger) *WeatherResult {
	cacheGetCtx, cacheGetCancel := context.WithTimeout(ctx, q.cacheTimeout)
	defer cacheGetCancel()

	previousWeather, previousExpiry, err := q.cache.Get(cacheGetCtx, resultCacheKey{})
	if err != nil {
		logger.Error(err, "Failed to retrieve result from cache.")
	}

	var result *WeatherResult

	if previousWeather != nil {
		cachedWeather := previousWeather.(resultCacheEntry) // Should never panic
		result = &WeatherResult{
			Weather:   cachedWeather.result,
			CreatedAt: cachedWeather.createdAt,
			Expiry:    previousExpiry,
		}

		logger.V(1).Info("Cache hit", "expires", previousExpiry.Sub(q.clock.Now()))
	}

	return result
}

// queryAllProvidersForWeather returns a weather summary by city name.
// It will query each provider one at a time until it gets a successful response to return.
func (q *Queryer) queryAllProvidersForWeather(
	ctx context.Context,
	logger logr.Logger,
	cityName string,
) (*weather.Summary, error) {
	for _, provider := range q.providers {
		res, err := q.queryProviderForWeather(ctx, cityName, provider)
		if err != nil {
			logger.Error(err, "Failed to query provider for weather.", providerLogKey, provider.ProviderName())
		}

		if res != nil {
			return res, nil
		}
	}

	return nil, errors.New("providerquery: no successful provider responses")
}

func (q *Queryer) queryProviderForWeather(
	ctx context.Context,
	cityName string,
	provider Provider,
) (*weather.Summary, error) {
	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), "providerquery: context done before exhausting providers")
	}

	ctx, cancel := context.WithTimeout(ctx, q.providerTimeout)
	defer cancel()

	weatherSummary, err := provider.GetWeatherSummary(ctx, cityName)
	if err != nil {
		return nil, errors.Wrap(err, "get weather summary")
	}

	return weatherSummary, nil
}

func (q *Queryer) cacheWeatherResult(ctx context.Context, logger logr.Logger, result *WeatherResult) {
	entry := resultCacheEntry{result: result.Weather, createdAt: result.CreatedAt}

	cacheSetCtx, cacheSetCancel := context.WithTimeout(ctx, q.cacheTimeout)
	defer cacheSetCancel()

	// Cache the new weather summary result.
	if err := q.cache.Set(cacheSetCtx, resultCacheKey{}, entry, result.Expiry); err != nil {
		logger.Error(err, "Failed to set result in cache.")
	} else {
		logger.V(1).Info("Cached result.")
	}
}
