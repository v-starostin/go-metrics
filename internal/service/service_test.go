package service_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/v-starostin/go-metrics/internal/mock"
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
