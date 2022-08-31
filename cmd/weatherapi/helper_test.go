package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// getOpenPort returns an open port.
func getOpenPort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}

	port := l.Addr().(*net.TCPAddr).Port

	err = l.Close()
	if err != nil {
		return 0, fmt.Errorf("close listener: %w", err)
	}

	return port, nil
}

// doHealthzRequest uses the default http client to hit the /healthz endpoint, returning
// the http response & error.
func doHealthzRequest(ctx context.Context, serverAddress string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/healthz", serverAddress), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create healthz request: %w", err)
	}

	return http.DefaultClient.Do(req)
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
