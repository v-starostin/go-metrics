// Code generated by mockery v2.43.2. DO NOT EDIT.

package mock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	model "github.com/v-starostin/go-metrics/internal/model"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// Load provides a mock function with given fields: ctx, mtype, mname
func (_m *Repository) Load(ctx context.Context, mtype string, mname string) (*model.Metric, error) {
	ret := _m.Called(ctx, mtype, mname)

	if len(ret) == 0 {
		panic("no return value specified for Load")
	}

	var r0 *model.Metric
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.Metric, error)); ok {
		return rf(ctx, mtype, mname)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Metric); ok {
		r0 = rf(ctx, mtype, mname)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Metric)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, mtype, mname)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadAll provides a mock function with given fields: ctx
func (_m *Repository) LoadAll(ctx context.Context) (model.Data, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for LoadAll")
	}

	var r0 model.Data
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (model.Data, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) model.Data); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.Data)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PingStorage provides a mock function with given fields: ctx
func (_m *Repository) PingStorage(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for PingStorage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RestoreFromFile provides a mock function with given fields:
func (_m *Repository) RestoreFromFile() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RestoreFromFile")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StoreMetric provides a mock function with given fields: ctx, m
func (_m *Repository) StoreMetric(ctx context.Context, m model.Metric) error {
	ret := _m.Called(ctx, m)

	if len(ret) == 0 {
		panic("no return value specified for StoreMetric")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.Metric) error); ok {
		r0 = rf(ctx, m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StoreMetrics provides a mock function with given fields: ctx, m
func (_m *Repository) StoreMetrics(ctx context.Context, m []model.Metric) error {
	ret := _m.Called(ctx, m)

	if len(ret) == 0 {
		panic("no return value specified for StoreMetrics")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []model.Metric) error); ok {
		r0 = rf(ctx, m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WriteToFile provides a mock function with given fields:
func (_m *Repository) WriteToFile() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for WriteToFile")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *Repository {
	mock := &Repository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
