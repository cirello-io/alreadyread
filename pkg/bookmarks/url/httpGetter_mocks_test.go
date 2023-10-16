// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package url

import (
	"net/http"
	"sync"
)

// Ensure, that httpGetterMock does implement httpGetter.
// If this is not the case, regenerate this file with moq.
var _ httpGetter = &httpGetterMock{}

// httpGetterMock is a mock implementation of httpGetter.
//
//	func TestSomethingThatUseshttpGetter(t *testing.T) {
//
//		// make and configure a mocked httpGetter
//		mockedhttpGetter := &httpGetterMock{
//			GetFunc: func(url string) (*http.Response, error) {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedhttpGetter in code that requires httpGetter
//		// and then make assertions.
//
//	}
type httpGetterMock struct {
	// GetFunc mocks the Get method.
	GetFunc func(url string) (*http.Response, error)

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// URL is the url argument value.
			URL string
		}
	}
	lockGet sync.RWMutex
}

// Get calls GetFunc.
func (mock *httpGetterMock) Get(url string) (*http.Response, error) {
	if mock.GetFunc == nil {
		panic("httpGetterMock.GetFunc: method is nil but httpGetter.Get was just called")
	}
	callInfo := struct {
		URL string
	}{
		URL: url,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(url)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedhttpGetter.GetCalls())
func (mock *httpGetterMock) GetCalls() []struct {
	URL string
} {
	var calls []struct {
		URL string
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}
