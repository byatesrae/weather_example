package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/byatesrae/weather/cmd/weatherapi/handlers"
	"github.com/byatesrae/weather/cmd/weatherapi/providers"
	"github.com/byatesrae/weather/internal/memorycache"
	"github.com/byatesrae/weather/internal/openweather"
	"github.com/byatesrae/weather/internal/providerquery"
	"github.com/byatesrae/weather/internal/weatherstack"
)

func main() {
	ctx := context.Background()
	logger := log.Default()

	c, err := loadConfig(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("[INF] Config: %+v.\n", c)

	server := createServer(c)

	go func() {
		logger.Printf("[INF] Server started, listening on %v.\n", server.Addr)

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[ERR] listen: %s\n", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERR] shutdown: %s\n", err)
		runtime.Goexit()
	}

	log.Print("[INF] Server Exited.\n")
}

func createServer(config *appConfig) *http.Server {
	providerHTTPClient := http.Client{Timeout: config.ResultTimeout}

	providerQueryer := providerquery.New(
		[]providerquery.Provider{
			providers.NewOpenWeatherProvider(
				openweather.New(
					config.OpenweatherEndpointURL,
					config.OpenweatherAPIKey,
					openweather.NewWithHTTPClient(&httpClientWithCorrelationID{innerClient: providerHTTPClient}),
				),
			),
			providers.NewWeatherStackProvider(
				weatherstack.New(
					config.WeatherstackEndpointURL,
					config.WeatherstackAccessKey,
					weatherstack.NewWithHTTPClient(&httpClientWithCorrelationID{innerClient: providerHTTPClient}),
				),
			),
		},
		memorycache.New(),
		providerquery.WithResultCacheTTL(config.ResultCacheTTL),
	)

	healthzHandler := handlers.NewHealthzHandler()
	weatherHandler := handlers.NewWeatherHandler(providerQueryer, config.ResultTimeout)

	router := mux.NewRouter().PathPrefix("/v1").Subrouter()
	router.Use(correlationIDMiddleware)
	router.Path("/healthz").Methods("GET").HandlerFunc(healthzHandler)
	router.Path("/weather").Methods("GET").Handler(weatherHandler)

	return &http.Server{
		Addr:              fmt.Sprintf(":%v", config.Port),
		Handler:           router,
		ReadHeaderTimeout: time.Second * 1,
	}
}

type correlationIDKey struct{}

// correlationIDMiddleware is middleware that adds a correlation ID to the context.
func correlationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		correlationID := req.Header.Get("X-Correlation-Id")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		req = req.WithContext(context.WithValue(req.Context(), correlationIDKey{}, correlationID))

		next.ServeHTTP(rw, req)
	})
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
		requestID := req.Context().Value(correlationIDKey{}).(string) // Should never panic

		if requestID != "" {
			req.Header.Set("X-Correlation-Id", requestID)
		}
	}

	return c.innerClient.Do(req)
}
