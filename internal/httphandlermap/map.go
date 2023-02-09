// Package httphandlermap contains a HTTP handler map that can be configured to
// change it's behaviour based on the request it receives.
package httphandlermap

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

// Map is an HTTP handler that can be configured to change it's behaviour based on
// the request it receives. It is safe for concurrent use.
type Map struct {
	// KeyGenFunc is used on receipt of a request to generate a "key" for the map.
	// The key is used to load a registered http Handler for invocation. See method
	// [Map.Register].
	//
	// The function takes a clone of the original request such that it might be used
	// in generating the key. The request clone body is always nil.
	KeyGenFunc func(*http.Request) any

	// DefaultHandler is used if [Map.KeyGenFunc] is nil or no handler is loaded.
	DefaultHandler http.HandlerFunc

	// Contains all registered handlers.
	handlerMap sync.Map
}

var _ http.Handler = (*Map)(nil)

// ServeHTTP loads a handler to serve the HTTP request.
func (m *Map) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.mustLoadHandler(req).ServeHTTP(rw, req)
}

// Register can be used to register a handler for a given key. Should that key be
// generated during [Map.ServeHTTP] then this handler will be invoked to serve the
// request. See field [Map.KeyGenFunc].
//
// The first value returned can be used to unregister the handler.
//
// Only one handler can be registered for any given key.
func (m *Map) Register(key any, h http.Handler) (func(), error) {
	if _, ok := m.handlerMap.Load(key); ok {
		return nil, fmt.Errorf("key %#v is already in use", key)
	}

	m.handlerMap.Store(key, h)

	return func() {
		m.handlerMap.Delete(key)
	}, nil
}

// mustLoadHandler loads a handler given a request.
func (m *Map) mustLoadHandler(req *http.Request) http.Handler {
	if m.KeyGenFunc == nil {
		if m.DefaultHandler != nil {
			return m.DefaultHandler
		}

		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}) // noop
	}

	key := m.KeyGenFunc(cloneRequest(req))

	loadedHandler, ok := m.handlerMap.Load(key)
	if !ok {
		if m.DefaultHandler != nil {
			return m.DefaultHandler
		}

		panic(fmt.Sprintf("No handler found for key \"%#v\".", key))
	}

	return loadedHandler.(http.Handler)
}

// cloneRequest is a helper function to clone a request (without keeping the body).
func cloneRequest(req *http.Request) *http.Request {
	reqCloneWithoutBody := req.Clone(context.Background())
	reqCloneWithoutBody.Body = nil

	return reqCloneWithoutBody
}
