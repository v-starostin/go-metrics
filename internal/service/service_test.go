package service_test

import (
	"testing"

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
	service := service.New(repo)
	suite.repo = repo
	suite.service = service
}

func TestService(t *testing.T) {
	suite.Run(t, new(serviceTestSuite))
}

func (suite *serviceTestSuite) TestMetric() {

	tt := []struct {
		name        string
		expected    *model.Metric
		metric      *model.Metric
		expectedErr string
	}{
		{
			name:     "good case",
			metric:   &model.Metric{Type: "gauge", Name: "metric2", Value: float64(1.23)},
			expected: &model.Metric{Type: "gauge", Name: "metric2", Value: float64(1.23)},
		},
		{
			name:        "bad case",
			metric:      &model.Metric{Type: "gauge", Name: "metric1", Value: float64(1.23)},
			expectedErr: "failed to load metric metric1",
		},
	}

	for _, test := range tt {
		suite.Run(test.name, func() {
			suite.repo.On("Load", test.metric.Type, test.metric.Name).Once().Return(test.expected)

			got, err := suite.service.Metric(test.metric.Type, test.metric.Name)
			if err != nil {
				suite.EqualError(err, test.expectedErr)
			} else {
				suite.Equal(test.expected, got)
			}
		})
	}
}

func (suite *serviceTestSuite) TestMetrics() {
	m1 := model.Metric{Type: "counter", Name: "metric1", Value: int64(10)}
	m2 := model.Metric{Type: "gauge", Name: "metric1", Value: float64(1.23)}
	m3 := model.Metric{Type: "gauge", Name: "metric2", Value: float64(1.24)}

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

			got, err := suite.service.Metrics()
			if err != nil {
				suite.EqualError(err, test.expectedErr)
			} else {
				suite.Equal(test.expected, got)
			}
		})
	}
}

func (suite *serviceTestSuite) TestServiceSave() {
	tt := []struct {
		name, mtype, mname, mvalue string
		expected                   error
		saved                      bool
	}{
		{
			name:     "good case (gauge)",
			mtype:    service.TypeGauge,
			mname:    "metric1",
			mvalue:   "2.0",
			saved:    true,
			expected: nil,
		},
		{
			name:     "good case (counter)",
			mtype:    service.TypeCounter,
			mname:    "metric1",
			mvalue:   "2",
			expected: nil,
			saved:    true,
		},
		{
			name:     "failed to parse float64",
			mtype:    service.TypeGauge,
			mname:    "metric1",
			mvalue:   "2,0",
			saved:    false,
			expected: service.ErrParseMetric,
		},
		{
			name:     "failed to parse int64",
			mtype:    service.TypeCounter,
			mname:    "metric1",
			mvalue:   "2,0",
			expected: service.ErrParseMetric,
		},

		{
			name:     "failed to store data (counter)",
			mtype:    service.TypeCounter,
			mname:    "metric1",
			mvalue:   "2",
			saved:    false,
			expected: service.ErrStoreData,
		},
		{
			name:     "failed to store data (gauge)",
			mtype:    service.TypeGauge,
			mname:    "metric1",
			mvalue:   "2.0",
			saved:    false,
			expected: service.ErrStoreData,
		},
	}

	for _, test := range tt {
		suite.Run(test.mname, func() {
			if test.mtype == service.TypeGauge {
				suite.repo.On("StoreGauge", test.mtype, test.mname, float64(2)).Once().Return(test.saved)
			}
			if test.mtype == service.TypeCounter {
				suite.repo.On("StoreCounter", test.mtype, test.mname, int64(2)).Once().Return(test.saved)
			}
			err := suite.service.Save(test.mtype, test.mname, test.mvalue)
			suite.Equal(test.expected, err)
		})
	}
}
