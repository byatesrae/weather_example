package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"

	"github.com/byatesrae/weather/internal/httphandlermap"
)

// serverURL is the URL of the server started with go main().
var serverURL string

// Test doubles.
var (
	// openweatherStubServerHandler is a test double shared by tests.
	openweatherStubServerHandler httphandlermap.Map

	// weatherstackStubServerHandler is a test double shared by tests.
	weatherstackStubServerHandler httphandlermap.Map
)

func TestMain(m *testing.M) {
	logger := newLogger(fmt.Sprintf("%s-component-test", component), false)

	ctx := context.Background()

	openweatherStubServerHandler := startOpenweatherStubServer(logger)
	weatherstackStubServerHandler := startWeatherstackStubServer(logger)

	config, err := newTestConfig(openweatherStubServerHandler.URL, weatherstackStubServerHandler.URL)
	if err != nil {
		logger.Error(err, "Failed to create test config, exiting.")
		os.Exit(1)
	}

	serverURL = fmt.Sprintf("http://127.0.0.1:%v", config.Port)

	originalArgs := os.Args
	os.Args = getMainArguments(config)
	stopMain := runMain()

	if err := verifyServerReady(ctx, logger, serverURL); err != nil {
		logger.Error(err, "Failed to verify main() has started, exiting.")
		os.Exit(1)
	}

	os.Args = originalArgs
	m.Run()

	weatherstackStubServerHandler.Close()
	openweatherStubServerHandler.Close()

	err = stopMain(ctx)
	if err != nil {
		logger.Error(err, "Failed to stop main, exiting.")
		os.Exit(1)
	}
}

// newTestConfig creates config that can be used in boostraping the server such that
// it can be tested.
func newTestConfig(openweatherURL, weatherstackURL string) (*appConfig, error) {
	serverPort, err := getOpenPort()
	if err != nil {
		return nil, fmt.Errorf("get open port for server: %w", err)
	}

	return &appConfig{
		Port:                    serverPort,
		OpenweatherEndpointURL:  openweatherURL,
		OpenweatherAPIKey:       "SET_BY_TESTMAIN",
		WeatherstackEndpointURL: weatherstackURL,
		WeatherstackAccessKey:   "SET_BY_TESTMAIN",
		ResultCacheTTL:          time.Millisecond * 500,
	}, nil
}

// getMainArguments will get all arguments required for main() to run successfully.
func getMainArguments(config *appConfig) []string {
	return []string{
		os.Args[0],
		fmt.Sprintf("-port=%v", config.Port),
		fmt.Sprintf("-openweather-endpoint-url=%s", config.OpenweatherEndpointURL),
		fmt.Sprintf("-openweather-api-key=%s", config.OpenweatherAPIKey),
		fmt.Sprintf("-weathertstack-endpoint-url=%s", config.WeatherstackEndpointURL),
		fmt.Sprintf("-weathertstack-access-key=%s", config.WeatherstackAccessKey),
		fmt.Sprintf("-result-cache-ttl=%s", config.ResultCacheTTL),
		fmt.Sprintf("-colourized-output=%v", config.ColourizedOutput),
	}
}

// runMain runs main() in a goroutine then blocks until the http API is ready to
// serve requests.
func runMain() func(ctx context.Context) error {
	mainDone := make(chan interface{})
	go func() {
		time.Sleep(time.Second * 3)

		main()

		mainDone <- nil
	}()

	return func(ctx context.Context) error {
		err := interruptThisProcess()
		if err != nil {
			return fmt.Errorf("interrupt this process: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		select {
		case <-mainDone:
		case <-ctx.Done():
			return fmt.Errorf("main() took too long to exit")
		}

		return nil
	}
}

// startWeatherstackStubServer starts an httptest.Server using the correlation ID
// header as a request discriminator.
func startWeatherstackStubServer(logger logr.Logger) *httptest.Server {
	weatherstackStubServerHandler.KeyGenFunc = getCorrelationIDOrNil

	s := httptest.NewServer(&weatherstackStubServerHandler)

	logger.V(0).Info("Weatherstack stub server started.", "addr", s.URL)

	return s
}

// startOpenweatherStubServer starts an httptest.Server using the correlation ID
// header as a request discriminator.
func startOpenweatherStubServer(logger logr.Logger) *httptest.Server {
	openweatherStubServerHandler.KeyGenFunc = getCorrelationIDOrNil

	s := httptest.NewServer(&openweatherStubServerHandler)

	logger.V(0).Info("Openweather stub server started.", "addr", s.URL)

	return s
}

func getCorrelationIDOrNil(r *http.Request) any {
	correlationID := r.Header.Get("X-Correlation-Id")

	if correlationID == "" {
		return nil
	}

	return correlationID
}

// getOpenPort returns an open port.
func getOpenPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port // Should never panic

	if err := listener.Close(); err != nil {
		return 0, fmt.Errorf("close listener: %w", err)
	}

	return port, nil
}

// doHealthzRequest uses the default http client to hit the /healthz endpoint, returning
// the http response & error.
func doHealthzRequest(ctx context.Context, serverURL string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/healthz", serverURL), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create healthz request: %w", err)
	}

	return http.DefaultClient.Do(req)
}

// verifyServerReady verifies that the server is ready to accept requests.
func verifyServerReady(ctx context.Context, logger logr.Logger, serverAddress string) error {
	var res *http.Response
	var resErr error

	for a := 0; a < 5; a++ {
		res, resErr = doHealthzRequest(ctx, serverAddress)
		if resErr == nil {
			break
		}

		logger.V(1).Info(fmt.Sprintf("Failed attempt %v to do healthz request.", a+1), "error_reason", resErr)

		time.Sleep(time.Second)
	}

	if resErr != nil {
		return fmt.Errorf("do healthz request: %w", resErr)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			logger.Error(err, "Failed to close healthz response body.")
		}
	}()

	return nil
}

// interruptThisProcess attempts to signal this process to be interrupted.
func interruptThisProcess() error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return fmt.Errorf("find this process: %w", err)
	}

	if err := p.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("send interrupt signal: %w", err)
	}

	return nil
}
