package otelmetrics

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/number"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

func TestMuxMiddleware(t *testing.T) {
	// Setup
	metricController := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
		controller.WithCollectPeriod(0),
	)

	mw, err := MuxMiddleware(metricController.Meter("Test123"))
	require.NoError(t, err, "create middleware")

	handler := func(rw http.ResponseWriter, _ *http.Request) {
		_, err := rw.Write([]byte{1, 2, 3, 4, 5})
		require.NoError(t, err, "write body")
	}

	router := mux.NewRouter()
	router.Use(mw)
	router.HandleFunc("/test/abc", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))

	// Do
	router.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/test/abc", http.NoBody),
	)

	// Assert
	err = metricController.Collect(context.Background())
	require.NoError(t, err, "collect metrics")

	expectedRecords := []testRecord{
		expectedRequestsInProgressRecord(http.MethodGet, "/test/abc", "HTTP/1.1", 0),
		expectedResponseSizeBytesRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusOK,
			aggregation.Buckets{
				Boundaries: []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1e+06, 2.5e+06, 5e+06, 1e+07},
				Counts:     []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			}),
		expectedRequestDurationSecondsRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusOK,
			aggregation.Buckets{
				Boundaries: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
				Counts:     []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			}),
		expectedRequestCountRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusOK, 1),
	}

	if !assertRecordsEqual(t, expectedRecords, getRecords(t, metricController)) {
		return
	}

	// Do (request with larger body)
	handler = func(rw http.ResponseWriter, _ *http.Request) {
		_, err := rw.Write(make([]byte, 1024*10))
		require.NoError(t, err, "write body")
	}
	router.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/test/abc", http.NoBody),
	)

	// Do (request takes longer)
	handler = func(rw http.ResponseWriter, _ *http.Request) {
		_, err := rw.Write(make([]byte, 1024))
		require.NoError(t, err, "write body")

		time.Sleep(time.Second)
	}
	router.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/test/abc", http.NoBody),
	)

	// Do (request with 500 error)
	handler = func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}
	router.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/test/abc", http.NoBody),
	)

	// Do (request that hangs)
	requestIsWaiting := make(chan bool)
	handler = func(_ http.ResponseWriter, req *http.Request) {
		requestIsWaiting <- true
		<-req.Context().Done()
	}
	waitCtx, waitCtxCancel := context.WithCancel(context.Background())
	t.Cleanup(waitCtxCancel)
	go router.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/test/abc", http.NoBody).Clone(waitCtx),
	)
	<-requestIsWaiting

	// Assert
	err = metricController.Collect(context.Background())
	require.NoError(t, err, "collect metrics")

	expectedRecords = []testRecord{
		expectedRequestsInProgressRecord(http.MethodGet, "/test/abc", "HTTP/1.1", 1),
		expectedResponseSizeBytesRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusOK,
			aggregation.Buckets{
				Boundaries: []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1e+06, 2.5e+06, 5e+06, 1e+07},
				Counts:     []uint64{2, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			}),
		expectedRequestDurationSecondsRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusOK,
			aggregation.Buckets{
				Boundaries: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
				Counts:     []uint64{2, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0},
			}),
		expectedRequestCountRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusOK, 3),
		expectedResponseSizeBytesRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusInternalServerError,
			aggregation.Buckets{
				Boundaries: []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1e+06, 2.5e+06, 5e+06, 1e+07},
				Counts:     []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			}),
		expectedRequestDurationSecondsRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusInternalServerError,
			aggregation.Buckets{
				Boundaries: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
				Counts:     []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			}),
		expectedRequestCountRecord(http.MethodGet, "/test/abc", "HTTP/1.1", http.StatusInternalServerError, 1),
	}

	if !assertRecordsEqual(t, expectedRecords, getRecords(t, metricController)) {
		return
	}
}

// testRecord is a summary of an otel metric export record.
type testRecord struct {
	name       string
	desc       string
	attributes []attribute.KeyValue

	// These are mutually exclusive (i.e if sum has a value, buckets & lastvalue will not.)
	buckets   aggregation.Buckets
	sum       number.Number
	lastValue number.Number
}

// getRecords retrieves all records (otel metric export records) from the metric
// controller.
func getRecords(t *testing.T, metricController *controller.Controller) []testRecord {
	t.Helper()

	testRecords := make([]testRecord, 0)

	err := metricController.ForEach(func(_ instrumentation.Library, exportReader export.Reader) error {
		return exportReader.ForEach(
			aggregation.CumulativeTemporalitySelector(),
			func(record export.Record) error {
				testRecord := testRecord{
					name: record.Descriptor().Name(),
					desc: record.Descriptor().Description(),
				}

				recordAggregation := record.Aggregation()

				attributes := record.Attributes()
				attributesIter := attributes.Iter()

				attributeKeyValues := make([]attribute.KeyValue, 0)
				for attributesIter.Next() {
					attributeKeyValues = append(attributeKeyValues, attributesIter.Attribute())
				}

				testRecord.attributes = attributeKeyValues

				switch v := recordAggregation.(type) {
				case aggregation.Histogram:
					buckets, err := v.Histogram()
					require.NoError(t, err, "get histogram buckets")

					testRecord.buckets = buckets
				case aggregation.Sum:
					num, err := v.Sum()
					require.NoError(t, err, "get sum")

					testRecord.sum = num

				case aggregation.LastValue:
					num, _, err := v.LastValue()
					require.NoError(t, err, "get last value")

					testRecord.lastValue = num

				default:
					t.Fatalf("unsupported aggregator: %s", recordAggregation.Kind())
				}

				testRecords = append(testRecords, testRecord)

				return nil
			})
	})
	require.NoError(t, err, "metricController.ForEach")

	return testRecords
}

// assertRecordsEqual asserts that expectedRecords equals actualRecords, with an
// ordering-insenstive comparison. Returns true if they are equal.
func assertRecordsEqual(t *testing.T, expectedRecords, actualRecords []testRecord) bool {
	t.Helper()

	if !assert.Len(t, actualRecords, len(expectedRecords), "actual records length") {
		return false
	}

	assertResult := true

	// Find a match between expected an actual
	for _, expectedRecord := range expectedRecords {
		var actualMatch *testRecord

		for a := range actualRecords {
			actualRecord := &actualRecords[a]

			if actualRecord.name == expectedRecord.name &&
				attributesEqual(actualRecord.attributes, expectedRecord.attributes) {
				actualMatch = actualRecord

				break
			}
		}

		if !assert.NotNil(t, actualMatch, "matching record") {
			assertResult = false

			continue
		}

		assertResult = assert.Equal(
			t,
			expectedRecord,
			*actualMatch,
			fmt.Sprintf("record %s (%v)", expectedRecord.name, expectedRecord.attributes),
		) && assertResult
	}

	return assertResult
}

// attributesEqual returns true if a = b (ordering-sensitive).
func attributesEqual(a, b []attribute.KeyValue) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// expectedRequestsInProgressRecord returns a testRecord that should match that
// derived from the RequestsInProgress metric.
func expectedRequestsInProgressRecord(method, path, proto string, lastValue int64) testRecord {
	return testRecord{
		name: DefaultRequestsInProgressName,
		desc: DefaultRequestsInProgressDesc,
		attributes: []attribute.KeyValue{
			{Key: DefaultRequestMethodAttributeKey, Value: attribute.StringValue(method)},
			{Key: DefaultRequestPathAttributeKey, Value: attribute.StringValue(path)},
			{Key: DefaultRequestProtoAttributeKey, Value: attribute.StringValue(proto)},
		},
		lastValue: number.NewInt64Number(lastValue),
	}
}

// expectedResponseSizeBytesRecord returns a testRecord that should match that
// derived from the ResponseSizeBytes metric.
func expectedResponseSizeBytesRecord(method, path, proto string, status int, buckets aggregation.Buckets) testRecord {
	return testRecord{
		name: DefaultResponseSizeBytesName,
		desc: DefaultResponseSizeBytesDesc,
		attributes: []attribute.KeyValue{
			{Key: DefaultRequestMethodAttributeKey, Value: attribute.StringValue(method)},
			{Key: DefaultRequestPathAttributeKey, Value: attribute.StringValue(path)},
			{Key: DefaultRequestProtoAttributeKey, Value: attribute.StringValue(proto)},
			{Key: DefaultResponseStatusAttributeKey, Value: attribute.IntValue(status)},
		},
		buckets: buckets,
	}
}

// expectedRequestDurationSecondsRecord returns a testRecord that should match that
// derived from the RequestDurationSeconds metric.
func expectedRequestDurationSecondsRecord(method, path, proto string, status int, buckets aggregation.Buckets) testRecord {
	return testRecord{
		name: DefaultRequestDurationSecondsName,
		desc: DefaultRequestDurationSecondsDesc,
		attributes: []attribute.KeyValue{
			{Key: DefaultRequestMethodAttributeKey, Value: attribute.StringValue(method)},
			{Key: DefaultRequestPathAttributeKey, Value: attribute.StringValue(path)},
			{Key: DefaultRequestProtoAttributeKey, Value: attribute.StringValue(proto)},
			{Key: DefaultResponseStatusAttributeKey, Value: attribute.IntValue(status)},
		},
		buckets: buckets,
	}
}

// expectedRequestCountRecord returns a testRecord that should match that
// derived from the RequestCount metric.
func expectedRequestCountRecord(method, path, proto string, status int, sum int64) testRecord {
	return testRecord{
		name: DefaultRequestCountName,
		desc: DefaultRequestCountDesc,
		attributes: []attribute.KeyValue{
			{Key: DefaultRequestMethodAttributeKey, Value: attribute.StringValue(method)},
			{Key: DefaultRequestPathAttributeKey, Value: attribute.StringValue(path)},
			{Key: DefaultRequestProtoAttributeKey, Value: attribute.StringValue(proto)},
			{Key: DefaultResponseStatusAttributeKey, Value: attribute.IntValue(status)},
		},
		sum: number.NewInt64Number(sum),
	}
}
