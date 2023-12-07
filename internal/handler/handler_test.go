package handler_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
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
	// m       *http.ServeMux
	r       *chi.Mux
	service *mock.Service
}

func (suite *handlerTestSuite) SetupTest() {
	service := &mock.Service{}
	h := handler.New(service)
	// m := http.NewServeMux()
	// m.Handle(updatePath, h)
	// m.Handle(upgradePath, h)
	// m.Handle(wrongPath, h)
	r := chi.NewRouter()
	r.Get("/", h.ServeHTTP)
	r.Get("/value/{type}/{name}", h.ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", h.ServeHTTP)
	suite.r = r
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
	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusOK, res.StatusCode)
	suite.Equal("metric metric1 of type gauge with value 1.23 has been set successfully", string(resBody))
}

func (suite *handlerTestSuite) TestHandlerServiceBadRequest() {
	req, err := http.NewRequest(http.MethodPost, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.service.On("Save", "gauge", "metric1", "1.23").Once().Return(service.ErrParseMetric)
	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusBadRequest, res.StatusCode)
	suite.Equal("bad request\n", string(resBody))
}

func (suite *handlerTestSuite) TestHandlerServiceError() {
	req, err := http.NewRequest(http.MethodPost, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.service.On("Save", "gauge", "metric1", "1.23").Once().Return(errors.New("err"))
	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusInternalServerError, res.StatusCode)
	suite.Equal("internal server error\n", string(resBody))
}

func (suite *handlerTestSuite) TestHandlerGetRequest() {
	req, err := http.NewRequest(http.MethodGet, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()

	suite.Equal(http.StatusMethodNotAllowed, res.StatusCode)
}

func (suite *handlerTestSuite) TestHandlerWrongCommand() {
	req, err := http.NewRequest(http.MethodPost, address+upgradePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusNotFound, res.StatusCode)
	suite.Equal("404 page not found\n", string(resBody))
}

func (suite *handlerTestSuite) TestHandlerEmptyURL() {
	req, err := http.NewRequest(http.MethodPost, address+wrongPath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusNotFound, res.StatusCode)
	suite.Equal("404 page not found\n", string(resBody))
}
