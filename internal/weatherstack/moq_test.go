// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package weatherstack

import (
	"net/http"
	"sync"
)

// Ensure, that HTTPClientMock does implement HTTPClient.
// If this is not the case, regenerate this file with moq.
var _ HTTPClient = &HTTPClientMock{}

// HTTPClientMock is a mock implementation of HTTPClient.
//
//	func TestSomethingThatUsesHTTPClient(t *testing.T) {
//
//		// make and configure a mocked HTTPClient
//		mockedHTTPClient := &HTTPClientMock{
//			DoFunc: func(req *http.Request) (*http.Response, error) {
//				panic("mock out the Do method")
//			},
//		}
//
//		// use mockedHTTPClient in code that requires HTTPClient
//		// and then make assertions.
//
//	}
type HTTPClientMock struct {
	// DoFunc mocks the Do method.
	DoFunc func(req *http.Request) (*http.Response, error)

	// calls tracks calls to the methods.
	calls struct {
		// Do holds details about calls to the Do method.
		Do []struct {
			// Req is the req argument value.
			Req *http.Request
		}
	}
	lockDo sync.RWMutex
}

// Do calls DoFunc.
func (mock *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	if mock.DoFunc == nil {
		panic("HTTPClientMock.DoFunc: method is nil but HTTPClient.Do was just called")
	}
	callInfo := struct {
		Req *http.Request
	}{
		Req: req,
	}
	mock.lockDo.Lock()
	mock.calls.Do = append(mock.calls.Do, callInfo)
	mock.lockDo.Unlock()
	return mock.DoFunc(req)
}

// DoCalls gets all the calls that were made to Do.
// Check the length with:
//
//	len(mockedHTTPClient.DoCalls())
func (mock *HTTPClientMock) DoCalls() []struct {
	Req *http.Request
} {
	var calls []struct {
		Req *http.Request
	}
	mock.lockDo.RLock()
	calls = mock.calls.Do
	mock.lockDo.RUnlock()
	return calls
}
