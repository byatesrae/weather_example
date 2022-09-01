package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	config, err := newTestConfig()
	if err != nil {
		log.Fatalf("[FTL] [TestMain] Failed to run build config: %v", err)
	}

	if err := setEnvVars(config); err != nil {
		log.Fatalf("[FTL] [TestMain] Failed to run set env vars: %v", err)
	}

	stopMain, err := runMain(ctx)
	if err != nil {
		log.Fatalf("[FTL] [TestMain] Failed to run main(): %v", err)
	}

	if err := verifyServerReady(ctx, fmt.Sprintf("http://127.0.0.1:%v", config.Port)); err != nil {
		log.Fatalf("[FTL] [TestMain] Failed to verify main() has started: %v", err)
	}

	m.Run()

	err = stopMain(ctx)
	if err != nil {
		log.Fatalf("[FTL] [TestMain] Failed to stop main(): %v", err)
	}
}

// newTestConfig creates config that can be used in boostraping the server such that
// it can be tested.
func newTestConfig() (*config, error) {
	serverPort, err := getOpenPort()
	if err != nil {
		return nil, fmt.Errorf("get open port for server: %w", err)
	}

	return &config{
		Port:                  serverPort,
		OpenweatherAPIKey:     "SET_BY_TESTMAIN",
		WeatherstackAccessKey: "SET_BY_TESTMAIN",
	}, nil
}

// setEnvVars will set all environment variables required for main() to run successfully.
func setEnvVars(c *config) error {
	if err := os.Setenv("OPENWEATHER_API_KEY", c.OpenweatherAPIKey); err != nil {
		return fmt.Errorf("set OPENWEATHER_API_KEY: %w", err)
	}

	if err := os.Setenv("WEATHERTSTACK_ACCESS_KEY", c.WeatherstackAccessKey); err != nil {
		return fmt.Errorf("set WEATHERTSTACK_ACCESS_KEY: %w", err)
	}

	if err := os.Setenv("PORT", fmt.Sprintf("%v", c.Port)); err != nil {
		return fmt.Errorf("set PORT: %w", err)
	}

	return nil
}

// runMain runs main() in a goroutine then blocks until the http API is ready to
// serve requests.
func runMain(ctx context.Context) (func(ctx context.Context) error, error) {
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
	}, nil
}

// verifyServerReady verifies that the server is ready to accept requests.
func verifyServerReady(ctx context.Context, serverAddress string) error {
	var res *http.Response
	var resErr error

	for a := 0; a < 5; a++ {
		res, resErr = doHealthzRequest(ctx, serverAddress, 0)
		if resErr == nil {
			break
		}

		log.Printf("[DBG] [TestMain] Failed attempt %v to do healthz request: %v", a+1, resErr)

		time.Sleep(time.Second)
	}

	if resErr != nil {
		return fmt.Errorf("do healthz request: %w", resErr)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Printf("[WAR] [TestMain] Failed to close healthz response body: %v", err)
		}
	}()

	return nil
}

func TestNoop(t *testing.T) {
	// Just have one empty test so TestMain runs/validates the server can start/stop.
}
