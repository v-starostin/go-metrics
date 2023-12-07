// Code generated by mockery v2.38.0. DO NOT EDIT.

package mock

import (
	mock "github.com/stretchr/testify/mock"
	model "github.com/v-starostin/go-metrics/internal/model"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Metric provides a mock function with given fields: mtype, mname
func (_m *Service) Metric(mtype string, mname string) (*model.Metric, error) {
	ret := _m.Called(mtype, mname)

	if len(ret) == 0 {
		panic("no return value specified for Metric")
	}

	var r0 *model.Metric
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (*model.Metric, error)); ok {
		return rf(mtype, mname)
	}
	if rf, ok := ret.Get(0).(func(string, string) *model.Metric); ok {
		r0 = rf(mtype, mname)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Metric)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(mtype, mname)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Metrics provides a mock function with given fields:
func (_m *Service) Metrics() (model.Data, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Metrics")
	}

	var r0 model.Data
	var r1 error
	if rf, ok := ret.Get(0).(func() (model.Data, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() model.Data); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.Data)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: mtype, mname, mvalue
func (_m *Service) Save(mtype string, mname string, mvalue string) error {
	ret := _m.Called(mtype, mname, mvalue)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(mtype, mname, mvalue)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
