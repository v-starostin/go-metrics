package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/v-starostin/go-metrics/internal/handler"
	"github.com/v-starostin/go-metrics/internal/mock"
	"github.com/v-starostin/go-metrics/internal/service"
)

const (
	address     = "http://0.0.0.0:8080"
	updatePath  = "/update/gauge/metric1/1.23"
	upgradePath = "/upgrade/gauge/metric1/1.23"
	wrongPath   = "/update/counter/metric2"
)

type handlerTestSuite struct {
	suite.Suite
	m       *http.ServeMux
	service *mock.Service
}

func (suite *handlerTestSuite) SetupTest() {
	service := &mock.Service{}
	h := handler.New(service)
	m := http.NewServeMux()
	m.Handle(updatePath, h)
	m.Handle(upgradePath, h)
	m.Handle(wrongPath, h)
	suite.m = m
	suite.service = service
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}

func (suite *handlerTestSuite) TestHandlerServiceOK() {
	req, err := http.NewRequest(http.MethodPost, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.service.On("Save", "gauge", "metric1", "1.23").Once().Return(nil)
	suite.m.ServeHTTP(rr, req)

	suite.Equal(http.StatusOK, rr.Result().StatusCode)
	suite.Equal("metric metric1 of type gauge with value 1.23 has been set successfully", rr.Body.String())
}

func (suite *handlerTestSuite) TestHandlerServiceBadRequest() {
	req, err := http.NewRequest(http.MethodPost, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.service.On("Save", "gauge", "metric1", "1.23").Once().Return(service.ErrParseMetric)
	suite.m.ServeHTTP(rr, req)

	suite.Equal(http.StatusBadRequest, rr.Result().StatusCode)
	suite.Equal("bad request\n", rr.Body.String())
}

func (suite *handlerTestSuite) TestHandlerServiceError() {
	req, err := http.NewRequest(http.MethodPost, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.service.On("Save", "gauge", "metric1", "1.23").Once().Return(errors.New("err"))
	suite.m.ServeHTTP(rr, req)

	suite.Equal(http.StatusInternalServerError, rr.Result().StatusCode)
	suite.Equal("service error\n", rr.Body.String())
}

func (suite *handlerTestSuite) TestHandlerGetRequest() {
	req, err := http.NewRequest(http.MethodGet, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.m.ServeHTTP(rr, req)

	suite.Equal(http.StatusBadRequest, rr.Result().StatusCode)
	suite.Equal("method GET is not supported\n", rr.Body.String())
}

func (suite *handlerTestSuite) TestHandlerWrongCommand() {
	req, err := http.NewRequest(http.MethodPost, address+upgradePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.m.ServeHTTP(rr, req)

	suite.Equal(http.StatusBadRequest, rr.Result().StatusCode)
	suite.Equal("bad request\n", rr.Body.String())
}

func (suite *handlerTestSuite) TestHandlerEmptyURL() {
	req, err := http.NewRequest(http.MethodPost, address+wrongPath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.m.ServeHTTP(rr, req)

	suite.Equal(http.StatusNotFound, rr.Result().StatusCode)
	suite.Equal("not found\n", rr.Body.String())
}
