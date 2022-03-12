package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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

	server := createServer(c)

	go func() {
		logger.Printf("[INF] Server started, listening on %v.\n", server.Addr)

		if err := server.ListenAndServe(); err != nil {
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

func createServer(c *config) *http.Server {
	providerQueryer := providerquery.New(
		[]providerquery.Provider{
			providers.NewOpenWeatherProvider(
				openweather.New(c.OpenweatherEndpointURL, c.OpenweatherAPIKey),
			),
			providers.NewWeatherStackProvider(
				weatherstack.New(c.WeatherstackEndpointURL, c.WeatherstackAccessKey),
			),
		},
		memorycache.New(),
	)

	healthzHandler := handlers.NewHealthzHandler()
	weatherHandler := handlers.NewWeatherHandler(providerQueryer, c.ResultTimeout)

	r := mux.NewRouter().PathPrefix("/v1").Subrouter()
	r.Path("/healthz").Methods("GET").HandlerFunc(healthzHandler)
	r.Path("/weather").Methods("GET").Handler(weatherHandler)

	return &http.Server{Addr: fmt.Sprintf(":%v", c.Port), Handler: r}
}
