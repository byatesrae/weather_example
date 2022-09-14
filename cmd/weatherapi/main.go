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
	hostmetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

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

	logger.Info("Config loaded.", "config", c.masked())

	prometheusExporter, err := instrument(component, "v0.0.0", "local")
	if err != nil {
		logger.Error(err, "Failed to instrument application.")
		os.Exit(1)
	}

	server := createServer(logger, c, prometheusExporter)

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

func createServer(
	logger logr.Logger,
	config *appConfig,
	prometheusExporter *prometheus.Exporter,
) *http.Server {
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

	rootRouter := mux.NewRouter()
	rootRouter.Path("/metrics").HandlerFunc(prometheusExporter.ServeHTTP)

	v1Router := rootRouter.PathPrefix("/v1").Subrouter()
	v1Router.Use(correlationIDMiddleware(logger))
	v1Router.Path("/healthz").Methods("GET").HandlerFunc(healthzHandler)
	v1Router.Path("/weather").Methods("GET").Handler(weatherHandler)

	return &http.Server{
		Addr:              fmt.Sprintf(":%v", config.Port),
		Handler:           rootRouter,
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

// instrument configures an otel metric controller.
func instrument(serviceName, serviceVersion, deploymentEnvironment string) (*prometheus.Exporter, error) {
	// Resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless([]attribute.KeyValue{
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.DeploymentEnvironmentKey.String(deploymentEnvironment),
		}...))
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	// Controller (metric provider)
	metricController := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(),
			aggregation.CumulativeTemporalitySelector(),
		),
		controller.WithResource(res),
	)

	// This is required for push exporters and optional for pull exporters. If Start()
	// is called then controller.ErrControllerStarted errors will start appearing
	// in the logs as the prometheus exporter tries to controller.Collect() when the
	// controller is already collecting.
	// metricController.Start(context.Background())

	if err := runtimemetrics.Start(runtimemetrics.WithMeterProvider(metricController)); err != nil {
		return nil, fmt.Errorf("start runtime metric gathering: %w", err)
	}

	if err := hostmetrics.Start(hostmetrics.WithMeterProvider(metricController)); err != nil {
		return nil, fmt.Errorf("start host metric gathering: %w", err)
	}

	// Prometheus exporter
	exporter, err := prometheus.New(
		prometheus.Config{},
		metricController)
	if err != nil {
		return nil, fmt.Errorf("create prometheus exporter: %w", err)
	}

	return exporter, nil
}
