// Code generated by mockery v2.38.0. DO NOT EDIT.

package mock

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// HTTPClient is an autogenerated mock type for the HTTPClient type
type HTTPClient struct {
	mock.Mock
}

// Do provides a mock function with given fields: req
func (_m *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	ret := _m.Called(req)

	if len(ret) == 0 {
		panic("no return value specified for Do")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(*http.Request) (*http.Response, error)); ok {
		return rf(req)
	}
	if rf, ok := ret.Get(0).(func(*http.Request) *http.Response); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(*http.Request) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewHTTPClient creates a new instance of HTTPClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHTTPClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *HTTPClient {
	mock := &HTTPClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
