// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package providerquery

import (
	"context"
	"github.com/byatesrae/weather"
	"sync"
	"time"
)

// Ensure, that ProviderMock does implement Provider.
// If this is not the case, regenerate this file with moq.
var _ Provider = &ProviderMock{}

// ProviderMock is a mock implementation of Provider.
//
//	func TestSomethingThatUsesProvider(t *testing.T) {
//
//		// make and configure a mocked Provider
//		mockedProvider := &ProviderMock{
//			GetWeatherSummaryFunc: func(ctx context.Context, cityName string) (*weather.Summary, error) {
//				panic("mock out the GetWeatherSummary method")
//			},
//			ProviderNameFunc: func() string {
//				panic("mock out the ProviderName method")
//			},
//		}
//
//		// use mockedProvider in code that requires Provider
//		// and then make assertions.
//
//	}
type ProviderMock struct {
	// GetWeatherSummaryFunc mocks the GetWeatherSummary method.
	GetWeatherSummaryFunc func(ctx context.Context, cityName string) (*weather.Summary, error)

	// ProviderNameFunc mocks the ProviderName method.
	ProviderNameFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// GetWeatherSummary holds details about calls to the GetWeatherSummary method.
		GetWeatherSummary []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// CityName is the cityName argument value.
			CityName string
		}
		// ProviderName holds details about calls to the ProviderName method.
		ProviderName []struct {
		}
	}
	lockGetWeatherSummary sync.RWMutex
	lockProviderName      sync.RWMutex
}

// GetWeatherSummary calls GetWeatherSummaryFunc.
func (mock *ProviderMock) GetWeatherSummary(ctx context.Context, cityName string) (*weather.Summary, error) {
	if mock.GetWeatherSummaryFunc == nil {
		panic("ProviderMock.GetWeatherSummaryFunc: method is nil but Provider.GetWeatherSummary was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		CityName string
	}{
		Ctx:      ctx,
		CityName: cityName,
	}
	mock.lockGetWeatherSummary.Lock()
	mock.calls.GetWeatherSummary = append(mock.calls.GetWeatherSummary, callInfo)
	mock.lockGetWeatherSummary.Unlock()
	return mock.GetWeatherSummaryFunc(ctx, cityName)
}

// GetWeatherSummaryCalls gets all the calls that were made to GetWeatherSummary.
// Check the length with:
//
//	len(mockedProvider.GetWeatherSummaryCalls())
func (mock *ProviderMock) GetWeatherSummaryCalls() []struct {
	Ctx      context.Context
	CityName string
} {
	var calls []struct {
		Ctx      context.Context
		CityName string
	}
	mock.lockGetWeatherSummary.RLock()
	calls = mock.calls.GetWeatherSummary
	mock.lockGetWeatherSummary.RUnlock()
	return calls
}

// ProviderName calls ProviderNameFunc.
func (mock *ProviderMock) ProviderName() string {
	if mock.ProviderNameFunc == nil {
		panic("ProviderMock.ProviderNameFunc: method is nil but Provider.ProviderName was just called")
	}
	callInfo := struct {
	}{}
	mock.lockProviderName.Lock()
	mock.calls.ProviderName = append(mock.calls.ProviderName, callInfo)
	mock.lockProviderName.Unlock()
	return mock.ProviderNameFunc()
}

// ProviderNameCalls gets all the calls that were made to ProviderName.
// Check the length with:
//
//	len(mockedProvider.ProviderNameCalls())
func (mock *ProviderMock) ProviderNameCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockProviderName.RLock()
	calls = mock.calls.ProviderName
	mock.lockProviderName.RUnlock()
	return calls
}

// Ensure, that CacheMock does implement Cache.
// If this is not the case, regenerate this file with moq.
var _ Cache = &CacheMock{}

// CacheMock is a mock implementation of Cache.
//
//	func TestSomethingThatUsesCache(t *testing.T) {
//
//		// make and configure a mocked Cache
//		mockedCache := &CacheMock{
//			GetFunc: func(ctx context.Context, key interface{}) (interface{}, time.Time, error) {
//				panic("mock out the Get method")
//			},
//			SetFunc: func(ctx context.Context, key interface{}, val interface{}, expiry time.Time) error {
//				panic("mock out the Set method")
//			},
//		}
//
//		// use mockedCache in code that requires Cache
//		// and then make assertions.
//
//	}
type CacheMock struct {
	// GetFunc mocks the Get method.
	GetFunc func(ctx context.Context, key interface{}) (interface{}, time.Time, error)

	// SetFunc mocks the Set method.
	SetFunc func(ctx context.Context, key interface{}, val interface{}, expiry time.Time) error

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Key is the key argument value.
			Key interface{}
		}
		// Set holds details about calls to the Set method.
		Set []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Key is the key argument value.
			Key interface{}
			// Val is the val argument value.
			Val interface{}
			// Expiry is the expiry argument value.
			Expiry time.Time
		}
	}
	lockGet sync.RWMutex
	lockSet sync.RWMutex
}

// Get calls GetFunc.
func (mock *CacheMock) Get(ctx context.Context, key interface{}) (interface{}, time.Time, error) {
	if mock.GetFunc == nil {
		panic("CacheMock.GetFunc: method is nil but Cache.Get was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Key interface{}
	}{
		Ctx: ctx,
		Key: key,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(ctx, key)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedCache.GetCalls())
func (mock *CacheMock) GetCalls() []struct {
	Ctx context.Context
	Key interface{}
} {
	var calls []struct {
		Ctx context.Context
		Key interface{}
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// Set calls SetFunc.
func (mock *CacheMock) Set(ctx context.Context, key interface{}, val interface{}, expiry time.Time) error {
	if mock.SetFunc == nil {
		panic("CacheMock.SetFunc: method is nil but Cache.Set was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Key    interface{}
		Val    interface{}
		Expiry time.Time
	}{
		Ctx:    ctx,
		Key:    key,
		Val:    val,
		Expiry: expiry,
	}
	mock.lockSet.Lock()
	mock.calls.Set = append(mock.calls.Set, callInfo)
	mock.lockSet.Unlock()
	return mock.SetFunc(ctx, key, val, expiry)
}

// SetCalls gets all the calls that were made to Set.
// Check the length with:
//
//	len(mockedCache.SetCalls())
func (mock *CacheMock) SetCalls() []struct {
	Ctx    context.Context
	Key    interface{}
	Val    interface{}
	Expiry time.Time
} {
	var calls []struct {
		Ctx    context.Context
		Key    interface{}
		Val    interface{}
		Expiry time.Time
	}
	mock.lockSet.RLock()
	calls = mock.calls.Set
	mock.lockSet.RUnlock()
	return calls
}
