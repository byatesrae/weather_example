package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/byatesrae/weather/cmd/weatherapi/handlers"
	"github.com/byatesrae/weather/cmd/weatherapi/providers"
	"github.com/byatesrae/weather/internal/memorycache"
	"github.com/byatesrae/weather/internal/openweather"
	"github.com/byatesrae/weather/internal/providerquery"
	"github.com/byatesrae/weather/internal/weatherstack"
)

const (
	component = "weather-api"
)

func main() {
	zerolog.SetGlobalLevel(-10)

	logger := newLogger().WithName(component)

	ctx := context.Background()
	ctx = setLoggerInContext(ctx, logger)

	c, err := loadConfig(ctx)
	if err != nil {
		logger.Error(err, "Failed to load config.")
		os.Exit(1)
	}

	logger.Info("Config loaded.", "config", c)

	server := createServer(logger, c)

	go func() {
		logger.Info("Server started.", "addr", server.Addr)

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err, "Error during listen/serving.")
			os.Exit(1)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error(err, "Error during shutdown.")
		runtime.Goexit()
	}

	logger.Info("Server exited.")
}

func newLogger() logr.Logger {
	zl := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339, NoColor: false})
	zl = zl.With().Timestamp().Logger().Level(-10)

	return zerologr.New(&zl)
}

func createServer(logger logr.Logger, config *appConfig) *http.Server {
	providerHTTPClient := http.Client{Timeout: config.ResultTimeout}

	providerQueryer := providerquery.New(
		[]providerquery.Provider{
			providers.NewOpenWeatherProvider(
				openweather.New(
					config.OpenweatherEndpointURL,
					config.OpenweatherAPIKey,
					openweather.NewWithHTTPClient(&httpClientWithCorrelationID{innerClient: providerHTTPClient}),
					openweather.WithGetLoggerFromContext(getLoggerFromContext),
				),
			),
			providers.NewWeatherStackProvider(
				weatherstack.New(
					config.WeatherstackEndpointURL,
					config.WeatherstackAccessKey,
					weatherstack.NewWithHTTPClient(&httpClientWithCorrelationID{innerClient: providerHTTPClient}),
					weatherstack.WithGetLoggerFromContext(getLoggerFromContext),
				),
			),
		},
		memorycache.New(),
		providerquery.WithResultCacheTTL(config.ResultCacheTTL),
		providerquery.WithGetLoggerFromContext(getLoggerFromContext),
	)

	healthzHandler := handlers.NewHealthzHandler(getLoggerFromContext)
	weatherHandler := handlers.NewWeatherHandler(providerQueryer, config.ResultTimeout, getLoggerFromContext)

	router := mux.NewRouter().PathPrefix("/v1").Subrouter()
	router.Use(correlationIDMiddleware(logger))
	router.Path("/healthz").Methods("GET").HandlerFunc(healthzHandler)
	router.Path("/weather").Methods("GET").Handler(weatherHandler)

	return &http.Server{
		Addr:              fmt.Sprintf(":%v", config.Port),
		Handler:           router,
		ReadHeaderTimeout: time.Second * 1,
	}
}

type correlationIDCtxKey struct{}

// correlationIDMiddleware is middleware that adds a correlation ID to the context.
func correlationIDMiddleware(logger logr.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			correlationID := req.Header.Get("X-Correlation-Id")
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			ctx := context.WithValue(req.Context(), correlationIDCtxKey{}, correlationID)
			ctx = setLoggerInContext(ctx, logger.WithValues("correlationID", correlationID))

			req = req.WithContext(ctx)

			next.ServeHTTP(rw, req)
		})
	}
}

type loggerCtxKey struct{}

// setLoggerInContext returns a copy of parent containing logger.
func setLoggerInContext(parent context.Context, logger logr.Logger) context.Context {
	return context.WithValue(parent, loggerCtxKey{}, logger)
}

// getLoggerFromContext will retrieve the logger from ctx that was set with setLoggerInContext().
func getLoggerFromContext(ctx context.Context) logr.Logger {
	return ctx.Value(loggerCtxKey{}).(logr.Logger) // should never panic
}

// httpClientWithCorrelationID is an http client that adds an "X-Correlation-Id"
// header to requests as they are sent. The value of this header is pulled out of
// the request context.
type httpClientWithCorrelationID struct {
	innerClient http.Client
}

// Do implements http.Client.Do, adding the "X-Correlation-Id" header (if possible)
// to the request.
func (c *httpClientWithCorrelationID) Do(req *http.Request) (*http.Response, error) {
	if _, ok := req.Header["X-Correlation-Id"]; !ok {
		requestID := req.Context().Value(correlationIDCtxKey{}).(string) // Should never panic

		if requestID != "" {
			req.Header.Set("X-Correlation-Id", requestID)
		}
	}

	return c.innerClient.Do(req)
}
