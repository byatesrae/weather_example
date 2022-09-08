package httphandlermap

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapServeHTTP(t *testing.T) {
	t.Parallel()

	t.Run("multiple_handlers", testMapServeHTTPWithMultipleHandlers)
	t.Run("map_zero_value", testMapServeHTTPWhenZeroValue)
	t.Run("default_no_key_gen", testMapServeHTTPCallsDefaultWithoutKeyGenFunc)
	t.Run("default_with_key_gen", testMapServeHTTPCallsDefaultWithKeyGenFunc)
}

func testMapServeHTTPWithMultipleHandlers(t *testing.T) {
	t.Parallel()

	callServeHTTP := func(m *Map, n int) {
		for a := 0; a < n; a++ {
			m.ServeHTTP(nil, &http.Request{})
		}
	}

	defaultHandler := &httphandlerMock{}

	handler1Key := "handler1"
	handler1 := &httphandlerMock{}
	expectHandler1CallCount := 2

	handler2Key := (*int)(nil)
	handler2 := &httphandlerMock{}
	expectHandler2CallCount := 5

	handler3Key := interface{}(nil)
	handler3 := &httphandlerMock{}
	expectHandler3CallCount := 8

	m := Map{
		DefaultHandler: defaultHandler.ServeHTTP,
	}

	_, err := m.Register(handler1Key, handler1)
	require.NoError(t, err, "register handler 1")

	_, err = m.Register(handler2Key, handler2)
	require.NoError(t, err, "register handler 2")

	_, err = m.Register(handler3Key, handler3)
	require.NoError(t, err, "register handler 3")

	m.KeyGenFunc = func(r *http.Request) any { return handler1Key }
	callServeHTTP(&m, expectHandler1CallCount)

	m.KeyGenFunc = func(r *http.Request) any { return handler2Key }
	callServeHTTP(&m, expectHandler2CallCount)

	m.KeyGenFunc = func(r *http.Request) any { return handler3Key }
	callServeHTTP(&m, expectHandler3CallCount)

	assert.Equal(t, defaultHandler.calledCount, 0, "default handler expected call count")
	assert.Equal(t, handler1.calledCount, expectHandler1CallCount, "handler 1 expected call count")
	assert.Equal(t, handler2.calledCount, expectHandler2CallCount, "handler 2 expected call count")
	assert.Equal(t, handler3.calledCount, expectHandler3CallCount, "handler 3 expected call count")
}

func testMapServeHTTPWhenZeroValue(t *testing.T) {
	t.Parallel()

	m := Map{}
	m.ServeHTTP(nil, nil)
}

func testMapServeHTTPCallsDefaultWithoutKeyGenFunc(t *testing.T) {
	t.Parallel()

	defaultHandler := &httphandlerMock{}
	m := Map{DefaultHandler: defaultHandler.ServeHTTP}

	m.ServeHTTP(nil, &http.Request{})

	assert.Equal(t, defaultHandler.calledCount, 1, "default handler call count")
}

func testMapServeHTTPCallsDefaultWithKeyGenFunc(t *testing.T) {
	t.Parallel()

	defaultHandler := &httphandlerMock{}
	m := Map{
		DefaultHandler: defaultHandler.ServeHTTP,
		KeyGenFunc:     func(r *http.Request) any { return "Test123" },
	}

	m.ServeHTTP(nil, &http.Request{})

	assert.Equal(t, defaultHandler.calledCount, 1, "default handler call count")
}

func TestMapRegister(t *testing.T) {
	t.Parallel()

	m := Map{}

	firstKey := "TestABC"
	firstCleanup, err := m.Register(firstKey, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	require.NoError(t, err, "first register error")
	require.NotNil(t, firstCleanup, "first cleanup function")

	secondCleanup, err := m.Register(firstKey, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	assert.EqualError(t, err, "key \"TestABC\" is already in use", "second register error")
	assert.Nil(t, secondCleanup, "second cleanup function")

	firstCleanup()

	thirdCleanup, err := m.Register(firstKey, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	require.NoError(t, err, "third register error")
	require.NotNil(t, thirdCleanup, "third cleanup function")
}

type httphandlerMock struct {
	calledCount int
}

func (hm *httphandlerMock) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	hm.calledCount++
}
