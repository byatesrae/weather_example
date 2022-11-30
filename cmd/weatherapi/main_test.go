package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/go-logr/logr"

	"github.com/byatesrae/weather/internal/httphandlermap"
)

// serverURL is the URL of the server.
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

	stopApp, err := runApp(ctx, config, logger)
	if err != nil {
		logger.Error(err, "Failed to run app, exiting.")
		os.Exit(1)
	}

	serverURL = fmt.Sprintf("http://127.0.0.1:%v", config.Port)

	if err := verifyServerReady(ctx, logger, serverURL); err != nil {
		logger.Error(err, "Failed to verify server is ready, exiting.")
		os.Exit(1)
	}

	m.Run()

	weatherstackStubServerHandler.Close()
	openweatherStubServerHandler.Close()

	err = stopApp(ctx)
	if err != nil {
		logger.Error(err, "Failed to stop application, exiting.")
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

func getRunAppCommand(ctx context.Context, appConfig *appConfig) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "go", "run", "./")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("OPENWEATHER_ENDPOINT_URL=%s", appConfig.OpenweatherEndpointURL),
		fmt.Sprintf("OPENWEATHER_API_KEY=%s", appConfig.OpenweatherAPIKey),
		fmt.Sprintf("WEATHERTSTACK_ENDPOINT_URL=%s", appConfig.WeatherstackEndpointURL),
		fmt.Sprintf("WEATHERSTACK_ACCESS_KEY=%s", appConfig.WeatherstackAccessKey),
		fmt.Sprintf("RESULT_CACHE_TTL=%s", appConfig.ResultCacheTTL.String()),
		fmt.Sprintf("PORT=%s", fmt.Sprintf("%v", appConfig.Port)),
	}...)

	//TODO
	//TODO
	//TODO
	//TODO
	//TODO
	// Dont reappend existing values

	return cmd
}

// runApp runs the application without blocking. The first return parameter can be
// used to stop the application.
func runApp(ctx context.Context, appConfig *appConfig, logger logr.Logger) (func(ctx context.Context) error, error) {
	cmd := getRunAppCommand(ctx, appConfig)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start app: %w", err)
	}

	cmdDone := make(chan interface{})
	go func() {
		if err := cmd.Wait(); err != nil {
			logger.Error(err, "error waiting for application to finish", "stdout", outb.String(), "stderr", errb.String())
		}

		cmdDone <- nil
	}()

	return func(ctx context.Context) error {
		p, err := os.FindProcess(-cmd.Process.Pid)
		if err != nil {
			return fmt.Errorf("find cmd process group: %w", err)
		}

		if err := p.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("send interrupt signal: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		select {
		case <-cmdDone:
		case <-ctx.Done():
			err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			if err != nil {
				return fmt.Errorf("Application took too long to exit gracefully and an attempt to kill it failed: %w", err)
			}

			return fmt.Errorf("Application took too long to exit gracefully, killed it instead.")
		}

		return nil
	}, nil
}

// startWeatherstackStubServer starts an httptest.Server using the correlation ID
// header as a request discriminator.
func startWeatherstackStubServer(logger logr.Logger) *httptest.Server {
	weatherstackStubServerHandler.KeyGenFunc = getCorrelationIDOrNil

	s := httptest.NewServer(&weatherstackStubServerHandler)

	logger.V(1).Info("Weatherstack stub server started.", "addr", s.URL)

	return s
}

// startOpenweatherStubServer starts an httptest.Server using the correlation ID
// header as a request discriminator.
func startOpenweatherStubServer(logger logr.Logger) *httptest.Server {
	openweatherStubServerHandler.KeyGenFunc = getCorrelationIDOrNil

	s := httptest.NewServer(&openweatherStubServerHandler)

	logger.V(1).Info("Openweather stub server started.", "addr", s.URL)

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
