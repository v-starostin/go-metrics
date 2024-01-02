package handler_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/v-starostin/go-metrics/internal/handler"
	"github.com/v-starostin/go-metrics/internal/mock"
	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

const (
	address         = "http://0.0.0.0:8080"
	updatePath      = "/update/gauge/metric1/1.23"
	getGaugePath    = "/value/gauge/metric1"
	getCounterPath  = "/value/counter/metric1"
	getAllPath      = "/"
	wrongMetricType = "/update/gauges/metric1/1.23"
)

type handlerTestSuite struct {
	suite.Suite
	r       *chi.Mux
	service *mock.Service
}

func (suite *handlerTestSuite) SetupTest() {
	srv := &mock.Service{}
	h := handler.New(&zerolog.Logger{}, srv)
	r := chi.NewRouter()
	r.Get("/", h.ServeHTTP)
	r.Get("/value/{type}/{name}", h.ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", h.ServeHTTP)
	suite.r = r
	suite.service = srv
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}

func (suite *handlerTestSuite) TestHandlerServiceOK() {
	req, err := http.NewRequest(http.MethodPost, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	f := new(float64)
	*f = 1.23

	m := model.Metric{MType: "gauge", ID: "metric1", Value: f}
	//suite.service.On("SaveMetric", "gauge", "metric1", "1.23").Once().Return(nil)
	suite.service.On("SaveMetric", m).Once().Return(nil)
	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusOK, res.StatusCode)
	suite.Equal(`"metric metric1 of type gauge with value 1.23 has been set successfully"`, string(resBody))
}

func (suite *handlerTestSuite) TestHandlerServiceBadRequest() {
	req, err := http.NewRequest(http.MethodPost, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	f := new(float64)
	*f = 1.23
	m := model.Metric{MType: "gauge", ID: "metric1", Value: f}
	suite.service.On("SaveMetric", m).Once().Return(service.ErrParseMetric)
	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusBadRequest, res.StatusCode)
	suite.Equal(`{"error":"Bad request"}`, string(resBody))
}

func (suite *handlerTestSuite) TestHandlerServiceError() {
	req, err := http.NewRequest(http.MethodPost, address+updatePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	f := new(float64)
	*f = 1.23
	m := model.Metric{MType: "gauge", ID: "metric1", Value: f}

	suite.service.On("SaveMetric", m).Once().Return(errors.New("err"))
	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusInternalServerError, res.StatusCode)
	suite.Equal(`{"error":"Internal server error"}`, string(resBody))
}

func (suite *handlerTestSuite) TestHandlerWrongMetricType() {
	req, err := http.NewRequest(http.MethodPost, address+wrongMetricType, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusBadRequest, res.StatusCode)
	suite.Equal(`{"error":"Bad request"}`, string(resBody))
}

func (suite *handlerTestSuite) TestHandlerGetGaugeOK() {
	req, err := http.NewRequest(http.MethodGet, address+getGaugePath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	f := new(float64)
	*f = 1.23

	m := &model.Metric{MType: "gauge", ID: "metric1", Value: f}
	suite.service.On("GetMetric", m.MType, m.ID).Once().Return(m, nil)

	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusOK, res.StatusCode)
	suite.Equal("1.23", string(resBody))
}

func (suite *handlerTestSuite) TestHandlerGetCounterOK() {
	req, err := http.NewRequest(http.MethodGet, address+getCounterPath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	i := new(int64)
	*i = 10

	m := &model.Metric{MType: "counter", ID: "metric1", Delta: i}
	suite.service.On("GetMetric", m.MType, m.ID).Once().Return(m, nil)

	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusOK, res.StatusCode)
	suite.Equal("10", string(resBody))
}

func (suite *handlerTestSuite) TestHandlerGetMetricNotFound() {
	req, err := http.NewRequest(http.MethodGet, address+getCounterPath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.service.On("GetMetric", "counter", "metric1").Once().Return(nil, errors.New("not found"))
	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusNotFound, res.StatusCode)
	suite.Equal(`{"error":"Not found"}`, string(resBody))
}

//func (suite *handlerTestSuite) TestHandlerGetAllOK() {
//	req, err := http.NewRequest(http.MethodGet, address+getAllPath, nil)
//	suite.NoError(err)
//
//	rr := httptest.NewRecorder()
//
//	f1, f2, i := new(float64), new(float64), new(int64)
//	*f1, *f2, *i = 1.23, 1.24, 10
//
//	m1 := model.Metric{MType: "counter", ID: "metric1", Delta: i}
//	m2 := model.Metric{MType: "gauge", ID: "metric1", Value: f1}
//	m3 := model.Metric{MType: "gauge", ID: "metric2", Value: f2}
//	d := model.Data(map[string]map[string]model.Metric{
//		"counter": {"metric1": m1},
//		"gauge":   {"metric1": m2, "metric2": m3},
//	})
//
//	suite.service.On("GetMetrics").Once().Return(d, nil)
//
//	suite.r.ServeHTTP(rr, req)
//	res := rr.Result()
//	defer res.Body.Close()
//	resBody, err := io.ReadAll(res.Body)
//	suite.NoError(err)
//
//	suite.Equal(http.StatusOK, res.StatusCode)
//	suite.Equal(expectedHTML, string(resBody))
//}

func (suite *handlerTestSuite) TestHandlerGetAllInternalServerError() {
	req, err := http.NewRequest(http.MethodGet, address+getAllPath, nil)
	suite.NoError(err)

	rr := httptest.NewRecorder()

	suite.service.On("GetMetrics").Once().Return(nil, errors.New("internal server error"))

	suite.r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	suite.NoError(err)

	suite.Equal(http.StatusInternalServerError, res.StatusCode)
	suite.Equal(`{"error":"Internal server error"}`, string(resBody))
}

var expectedHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    <ul>
    
        <li>metric1: 10</li>
    
        <li>metric1: 1.23</li>
    
        <li>metric2: 1.24</li>
    
    </ul>
</body>
</html>
`
