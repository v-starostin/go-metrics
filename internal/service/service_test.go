package service_test

import (
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
			expectedErr: "failed to load metric metric1",
		},
	}

	for _, test := range tt {
		suite.Run(test.name, func() {
			suite.repo.On("Load", test.metric.MType, test.metric.ID).Once().Return(test.expected)

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
		expectedErr string
		expected    model.Data
	}{
		{
			name:     "good case",
			expected: data,
		},
		{
			name:        "bad case",
			expectedErr: "failed to load metrics",
		},
	}

	for _, test := range tt {
		suite.Run(test.name, func() {
			suite.repo.On("LoadAll").Once().Return(test.expected)

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
		expected error
		saved    bool
	}{
		{
			name: "good case (gauge)",
			m: model.Metric{
				MType: service.TypeGauge,
				ID:    "metric1",
				Value: f1,
			},
			saved:    true,
			expected: nil,
		},
		{
			name: "good case (counter)",
			m: model.Metric{
				MType: service.TypeCounter,
				ID:    "metric1",
				Delta: i,
			},
			expected: nil,
			saved:    true,
		},
		{
			name: "failed to store data (counter)",
			m: model.Metric{
				MType: service.TypeCounter,
				ID:    "metric1",
				Delta: i,
			},
			saved:    false,
			expected: service.ErrStoreData,
		},
		{
			name: "failed to store data (gauge)",
			m: model.Metric{
				MType: service.TypeGauge,
				ID:    "metric1",
				Value: f1,
			},
			saved:    false,
			expected: service.ErrStoreData,
		},
	}

	for _, test := range tt {
		suite.Run(test.name, func() {
			suite.repo.On("Store", test.m).Once().Return(test.saved)

			err := suite.service.SaveMetric(test.m)
			suite.Equal(test.expected, err)
		})
	}
}
