package otelmetrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
)

const (
	// DefaultResponseSizeBytesName is the default name of the metric used to record
	// the size of outgoing responses in bytes.
	DefaultResponseSizeBytesName = "response_size_bytes"

	// DefaultResponseSizeBytesDesc is the default description of the metric used
	// to record the size of outgoing responses in bytes.
	DefaultResponseSizeBytesDesc = "The size of outgoing responses in bytes."

	// DefaultRequestDurationSecondsName is the default name of the metric used to record
	// the duration of incoming requests in seconds.
	DefaultRequestDurationSecondsName = "request_duration_seconds"

	// DefaultRequestDurationSecondsDesc is the default description of the metric used
	// to record the duration of incoming requests in seconds.
	DefaultRequestDurationSecondsDesc = "The duration of incoming requests in seconds."

	// DefaultRequestCountName is the default description of the metric used
	// to record the number of incoming requests.
	DefaultRequestCountName = "request_count"

	// DefaultRequestCountDesc is the default description of the metric used
	// to record the number of incoming requests.
	DefaultRequestCountDesc = "The number of incoming requests."

	// DefaultRequestsInProgressName is the default description of the metric used
	// to record the number of requests in progress.
	DefaultRequestsInProgressName = "requests_in_progress"

	// DefaultRequestsInProgressDesc is the default description of the metric used
	// to record the number of requests in progress.
	DefaultRequestsInProgressDesc = "The number of requests in progress."

	// DefaultRequestProtoAttributeKey will be the key used to attach to metrics
	// the value of the incoming request's http.Request.Proto field.
	DefaultRequestProtoAttributeKey = "proto"

	// DefaultResponseStatusAttributeKey will be the key used to attach to metrics
	// the value of the outgoing response's http.Response.StatusCode field.
	DefaultResponseStatusAttributeKey = "status"

	// DefaultRequestMethodAttributeKey will be the key used to attach to metrics
	// the value of the incoming request's http.Request.Method field.
	DefaultRequestMethodAttributeKey = "method"

	// DefaultRequestPathAttributeKey will be the key used to attach to metrics
	// the value of the incoming request's matched Mux route path template.
	DefaultRequestPathAttributeKey = "path"
)

// MuxMiddleware will capture metrics using meter and can be used as a gorilla/mux
// middleware. Metrics include:
//   - response_size_bytes
//   - request_duration_seconds
//   - request_count
//   - requests_in_progress
func MuxMiddleware(meter metric.Meter) (mux.MiddlewareFunc, error) {
	responseSizeBytes, err := meter.SyncInt64().Histogram(
		DefaultResponseSizeBytesName,
		instrument.WithDescription(DefaultResponseSizeBytesDesc),
		instrument.WithUnit("B"),
	)
	if err != nil {
		return nil, fmt.Errorf("create response_size_bytes metric: %w", err)
	}

	requestDurationSeconds, err := meter.SyncFloat64().Histogram(
		DefaultRequestDurationSecondsName,
		instrument.WithDescription(DefaultRequestDurationSecondsDesc),
		instrument.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("create request_duration_seconds metric: %w", err)
	}

	requestCount, err := meter.SyncInt64().Counter(
		DefaultRequestCountName,
		instrument.WithDescription(DefaultRequestCountDesc),
	)
	if err != nil {
		return nil, fmt.Errorf("create request_count metric: %w", err)
	}

	requestsInProgress, err := meter.AsyncInt64().Gauge(
		DefaultRequestsInProgressName,
		instrument.WithDescription(DefaultRequestsInProgressDesc),
	)
	if err != nil {
		return nil, fmt.Errorf("create requests_in_progress metric: %w", err)
	}

	requestCounter := newRequestCounter()
	err = meter.RegisterCallback([]instrument.Asynchronous{requestsInProgress}, func(ctx context.Context) {
		requestCounts := requestCounter.Get()

		for k, v := range requestCounts {
			requestsInProgress.Observe(ctx, int64(v), []attribute.KeyValue{
				attribute.String(DefaultRequestProtoAttributeKey, k.proto),
				attribute.String(DefaultRequestMethodAttributeKey, k.method),
				attribute.String(DefaultRequestPathAttributeKey, k.path),
			}...)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("register callback for requests_in_progress metric: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			path, _ := mux.CurrentRoute(req).GetPathTemplate()

			requestCounterKey := requestCounterKey{
				proto:  req.Proto,
				method: req.Method,
				path:   path,
			}
			requestCounter.Add(requestCounterKey, 1)
			defer func() {
				requestCounter.Add(requestCounterKey, -1)
			}()

			rwr := rwRecorder{ResponseWriter: rw}

			start := time.Now()
			next.ServeHTTP(&rwr, req)
			duration := time.Since(start)

			attributes := []attribute.KeyValue{
				attribute.String(DefaultRequestProtoAttributeKey, req.Proto),
				attribute.Int(DefaultResponseStatusAttributeKey, rwr.statusCode),
				attribute.String(DefaultRequestMethodAttributeKey, req.Method),
				attribute.String(DefaultRequestPathAttributeKey, path),
			}

			responseSizeBytes.Record(context.Background(), rwr.count, attributes...)
			requestDurationSeconds.Record(context.Background(), duration.Seconds(), attributes...)
			requestCount.Add(context.Background(), 1, attributes...)
		})
	}, nil
}

// rwRecorder is implements http.ResponseWriter while capturing various metrics.
type rwRecorder struct {
	http.ResponseWriter
	statusCode int   // The value of the last call to http.ResponseWriter.WriteHeader().
	count      int64 // The number of bytes written with http.ResponseWriter.Write().
}

func (rw *rwRecorder) Write(b []byte) (int, error) {
	rw.statusCode = http.StatusOK

	n, err := rw.ResponseWriter.Write(b)
	rw.count += int64(n)

	return n, err
}

func (rw *rwRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// requestCounterKey is used as a key for requestCounter.
type requestCounterKey struct {
	proto  string
	method string
	path   string
}

// requestCounter is a simple request counter safe for use across goroutines.
type requestCounter struct {
	countMap map[requestCounterKey]int
	mu       sync.RWMutex
}

func newRequestCounter() *requestCounter {
	return &requestCounter{
		countMap: make(map[requestCounterKey]int),
	}
}

// Add increments by i the count for request represented by key.
func (r *requestCounter) Add(key requestCounterKey, i int) {
	r.mu.Lock()

	r.countMap[key] += i

	r.mu.Unlock()
}

// Get retrieves the count of each key.
func (r *requestCounter) Get() map[requestCounterKey]int {
	mapCopy := make(map[requestCounterKey]int)

	r.mu.RLock()

	for k, v := range r.countMap {
		mapCopy[k] = v
	}

	r.mu.RUnlock()

	return mapCopy
}
