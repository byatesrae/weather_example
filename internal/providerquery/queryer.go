package providerquery

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"

	"github.com/byatesrae/weather"
)

// WeatherResult contains a weather summary and timeline data.
type WeatherResult struct {
	Expiry    time.Time
	CreatedAt time.Time
	Weather   *weather.Summary
}

// resultCacheKey is used as a key to cache resultCacheEntry.
type resultCacheKey struct{}

// resultCacheEntry wraps a weather summary to be cached.
type resultCacheEntry struct {
	result    *weather.Summary
	createdAt time.Time
}

// Cache is used to store & retrieve responses.
type Cache interface {
	Get(ctx context.Context, key interface{}) (interface{}, time.Time, error)
	Set(ctx context.Context, key, val interface{}, expiry time.Time) error
}

// Queryer will query a list of providers for a weather summary.
type Queryer struct {
	logger *log.Logger
	cache  Cache

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

// newOptions are options for the New function.
type newOptions struct {
	clock          Clock
	resultCacheTTL time.Duration
}

// withClock sets the clock used in the New function.
func withClock(clock Clock) func(o *newOptions) {
	return func(o *newOptions) {
		o.clock = clock
	}
}

// WithResultCacheTTL sets the amount of time a result is cached for.
func WithResultCacheTTL(resultCacheTTL time.Duration) func(o *newOptions) {
	return func(o *newOptions) {
		o.resultCacheTTL = resultCacheTTL
	}
}

// New creates a new Queryer.
func New(providers []Provider, cache Cache, overrides ...func(o *newOptions)) *Queryer {
	options := &newOptions{
		clock:          standardClock{},
		resultCacheTTL: time.Second * 3,
	}

	for _, override := range overrides {
		override(options)
	}

	return &Queryer{
		logger:          log.Default(),
		cache:           cache,
		cacheTimeout:    time.Second * 2, // These timeouts should all be configurable.
		providers:       providers,
		providerTimeout: time.Second * 3,
		resultCacheTTL:  options.resultCacheTTL,
		resultTimeout:   time.Second * 10,
		clock:           options.clock,
	}
}

// ReadWeatherResult will query one or more providers for a weather result. The result
// will be cached and sometimes served stale.
func (q *Queryer) ReadWeatherResult(ctx context.Context, city string) (*WeatherResult, error) {
	result := q.getCachedReadWeatherResult(ctx, city)

	retrievedCachedResult := result != nil

	if !retrievedCachedResult || q.clock.Now().After(result.Expiry) {
		log.Printf("[DBG] Querying all providers")

		queryAllProvidersCtx, queryAllProvidersCancel := context.WithTimeout(ctx, q.resultTimeout)
		defer queryAllProvidersCancel()

		newWeather, err, _ := q.queryAllProvidersForWeatherOnce.Do("queryAllProvidersForWeather", func() (interface{}, error) {
			return q.queryAllProvidersForWeather(queryAllProvidersCtx, city)
		})
		if err != nil {
			q.logger.Printf("[ERR] Failed to retrieve new weather: %s\n", err)

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

			go q.cacheWeatherResult(ctx, result)
		}
	}

	return result, nil
}

func (q *Queryer) getCachedReadWeatherResult(ctx context.Context, _ string) *WeatherResult {
	cacheGetCtx, cacheGetCancel := context.WithTimeout(ctx, q.cacheTimeout)
	defer cacheGetCancel()

	previousWeather, previousExpiry, err := q.cache.Get(cacheGetCtx, resultCacheKey{})
	if err != nil {
		q.logger.Printf("[ERR] Failed to retrieve result from cache: %s\n", err)
	}

	var result *WeatherResult

	if previousWeather != nil {
		cachedWeather := previousWeather.(resultCacheEntry) // Should never panic
		result = &WeatherResult{
			Weather:   cachedWeather.result,
			CreatedAt: cachedWeather.createdAt,
			Expiry:    previousExpiry,
		}

		log.Printf("[DBG] Cache hit, expires: %v\n", previousExpiry.Sub(q.clock.Now()))
	}

	return result
}

// queryAllProvidersForWeather returns a weather summary by city name.
// It will query each provider one at a time until it gets a successful response to return.
func (q *Queryer) queryAllProvidersForWeather(ctx context.Context, cityName string) (*weather.Summary, error) {
	for _, provider := range q.providers {
		res, err := q.queryProviderForWeather(ctx, cityName, provider)
		if err != nil {
			q.logger.Printf("providerquery: provider \"%s\" responded with err: %s\n", provider.ProviderName(), err)
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

func (q *Queryer) cacheWeatherResult(ctx context.Context, result *WeatherResult) {
	entry := resultCacheEntry{result: result.Weather, createdAt: result.CreatedAt}

	cacheSetCtx, cacheSetCancel := context.WithTimeout(ctx, q.cacheTimeout)
	defer cacheSetCancel()

	// Cache the new weather summary result.
	if err := q.cache.Set(cacheSetCtx, resultCacheKey{}, entry, result.Expiry); err != nil {
		q.logger.Printf("[ERR] Failed to set result in cache: %s\n", err)
	} else {
		log.Print("[DBG] Cached result.\n")
	}
}
