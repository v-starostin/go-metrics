package service_test

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/v-starostin/go-metrics/internal/mock"
	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

type serviceTestSuite struct {
	suite.Suite
	service *service.Service
	repo    *mock.Repository
}

func (suite *serviceTestSuite) SetupTest() {
	repo := &mock.Repository{}
	srv := service.New(&zerolog.Logger{}, repo)
	suite.repo = repo
	suite.service = srv
}

func TestService(t *testing.T) {
	suite.Run(t, new(serviceTestSuite))
}

func (suite *serviceTestSuite) TestMetric() {
	f := new(float64)
	*f = 1.23

	tt := []struct {
		name        string
		expected    *model.Metric
		metric      *model.Metric
		err         error
		expectedErr string
	}{
		{
			name:     "good case",
			metric:   &model.Metric{MType: "gauge", ID: "metric2", Value: f},
			expected: &model.Metric{MType: "gauge", ID: "metric2", Value: f},
		},
		{
			name:        "bad case",
			metric:      &model.Metric{MType: "gauge", ID: "metric1", Value: f},
			err:         errors.New("err"),
			expectedErr: "failed to load metric metric1: err",
		},
	}

	for _, test := range tt {
		suite.Run(test.name, func() {
			suite.repo.On("Load", test.metric.MType, test.metric.ID).Once().Return(test.metric, test.err)

			got, err := suite.service.GetMetric(test.metric.MType, test.metric.ID)
			if err != nil {
				suite.EqualError(err, test.expectedErr)
			} else {
				suite.Equal(test.expected, got)
			}
		})
	}
}

func (suite *serviceTestSuite) TestMetrics() {
	f1, f2, i := new(float64), new(float64), new(int64)
	*f1, *f2, *i = 1.23, 1.24, 10
	m1 := model.Metric{MType: "counter", ID: "metric1", Delta: i}
	m2 := model.Metric{MType: "gauge", ID: "metric1", Value: f1}
	m3 := model.Metric{MType: "gauge", ID: "metric2", Value: f2}

	data := model.Data(map[string]map[string]model.Metric{
		"counter": {"metric1": m1},
		"gauge":   {"metric1": m2, "metric2": m3},
	})

	tt := []struct {
		name        string
		err         error
		expectedErr string
		expected    model.Data
	}{
		{
			name:     "good case",
			expected: data,
		},
		{
			name:        "bad case",
			err:         errors.New("err"),
			expectedErr: "failed to load metrics: err",
		},
	}

	for _, test := range tt {
		suite.Run(test.name, func() {
			suite.repo.On("LoadAll").Once().Return(test.expected, test.err)

			got, err := suite.service.GetMetrics()
			if err != nil {
				suite.EqualError(err, test.expectedErr)
			} else {
				suite.Equal(test.expected, got)
			}
		})
	}
}

func (suite *serviceTestSuite) TestServiceSave() {
	f1, i := new(float64), new(int64)
	*f1, *i = 2.0, 2
	tt := []struct {
		name     string
		m        model.Metric
		expected string
		err      error
	}{
		{
			name: "good case (gauge)",
			m: model.Metric{
				MType: service.TypeGauge,
				ID:    "metric1",
				Value: f1,
			},
		},
		{
			name: "good case (counter)",
			m: model.Metric{
				MType: service.TypeCounter,
				ID:    "metric1",
				Delta: i,
			},
		},
		{
			name: "failed to store data (counter)",
			m: model.Metric{
				MType: service.TypeCounter,
				ID:    "metric1",
				Delta: i,
			},
			err:      errors.New("err"),
			expected: "failed to store data: err",
		},
		{
			name: "failed to store data (gauge)",
			m: model.Metric{
				MType: service.TypeGauge,
				ID:    "metric1",
				Value: f1,
			},
			err:      errors.New("err"),
			expected: "failed to store data: err",
		},
	}

	for _, test := range tt {
		suite.Run(test.name, func() {
			suite.repo.On("Store", test.m).Once().Return(test.err)

			err := suite.service.SaveMetric(test.m)
			if test.expected != "" {
				suite.EqualError(err, test.expected)
			} else {
				suite.NoError(err)
			}
		})
	}
}
