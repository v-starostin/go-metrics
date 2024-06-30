package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	mmock "github.com/stretchr/testify/mock"

	"github.com/v-starostin/go-metrics/internal/handler"
	"github.com/v-starostin/go-metrics/internal/mock"
	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

func setupRouter() (*chi.Mux, *mock.Service) {
	l := zerolog.New(zerolog.ConsoleWriter{Out: bytes.NewBuffer(nil)})
	srv := &mock.Service{}

	getMetricHandler := handler.NewGetMetric(&l, srv, key)
	getMetricsHandler := handler.NewGetMetrics(&l, srv, key)
	postMetricHandler := handler.NewPostMetric(&l, srv)
	postMetricsHandler := handler.NewPostMetrics(&l, srv, nil)
	getMetricV2Handler := handler.NewGetMetricV2(&l, srv, key)
	postMetricV2Handler := handler.NewPostMetricV2(&l, srv)
	pingStorageHandler := handler.NewPingStorage(&l, srv)

	r := chi.NewRouter()
	r.Get("/", getMetricsHandler.ServeHTTP)
	r.Get("/ping", pingStorageHandler.ServeHTTP)
	r.Get("/value/{type}/{name}", getMetricHandler.ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", postMetricHandler.ServeHTTP)
	r.Post("/value/", getMetricV2Handler.ServeHTTP)
	r.Post("/update/", postMetricV2Handler.ServeHTTP)
	r.Post("/updates/", postMetricsHandler.ServeHTTP)

	return r, srv
}

func ExamplePostMetric_ServeHTTP_ok() {
	r, srv := setupRouter()
	req, _ := http.NewRequest(http.MethodPost, address+updatePathCounter, nil)
	rr := httptest.NewRecorder()

	f := int64(1)
	m := model.Metric{MType: "counter", ID: "metric1", Delta: &f}
	srv.On("SaveMetric", m).Once().Return(nil)

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 200
	// "metric metric1 of type counter with value 1 has been set successfully"
}

func ExamplePostMetric_ServeHTTP_badRequest() {
	r, srv := setupRouter()
	req, _ := http.NewRequest(http.MethodPost, address+updatePathGauge, nil)
	rr := httptest.NewRecorder()

	f := 1.23
	m := model.Metric{MType: "gauge", ID: "metric1", Value: &f}
	srv.On("SaveMetric", m).Return(service.ErrParseMetric).Once()

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 400
	// {"error":"Bad request"}
}

func ExamplePostMetric_ServeHTTP_error() {
	r, srv := setupRouter()
	req, _ := http.NewRequest(http.MethodPost, address+updatePathGauge, nil)
	rr := httptest.NewRecorder()

	f := 1.23
	m := model.Metric{MType: "gauge", ID: "metric1", Value: &f}
	srv.On("SaveMetric", m).Return(errors.New("err")).Once()

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 500
	// {"error":"Internal server error"}
}

func ExamplePostMetric_ServeHTTP_wrongMetricType() {
	r, _ := setupRouter()
	req, _ := http.NewRequest(http.MethodPost, address+wrongMetricType, nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 400
	// {"error":"Bad request"}
}

func ExamplePostMetrics_ServeHTTP_ok() {
	r, srv := setupRouter()

	f1, f2 := new(float64), new(float64)
	*f1, *f2 = 1.25, 2.45
	m := []model.Metric{
		{MType: "gauge", ID: "metric1", Value: f1},
		{MType: "gauge", ID: "metric2", Value: f2},
	}
	b, _ := json.Marshal(m)
	req, _ := http.NewRequest(http.MethodPost, address+postMetrics, bytes.NewReader(b))
	rr := httptest.NewRecorder()
	srv.On("SaveMetrics", m).Once().Return(nil)

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 200
	// [{"id":"metric1","type":"gauge","value":1.25},{"id":"metric2","type":"gauge","value":2.45}]
}

func ExamplePostMetrics_ServeHTTP_serviceError() {
	r, srv := setupRouter()

	f1, f2 := new(float64), new(float64)
	*f1, *f2 = 1.25, 2.45
	m := []model.Metric{
		{MType: "gauge", ID: "metric1", Value: f1},
		{MType: "gauge", ID: "metric2", Value: f2},
	}
	b, _ := json.Marshal(m)
	req, _ := http.NewRequest(http.MethodPost, address+postMetrics, bytes.NewReader(b))
	rr := httptest.NewRecorder()
	srv.On("SaveMetrics", m).Once().Return(errors.New("service error"))

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 500
	// {"error":"Internal server error"}
}

func ExamplePostMetrics_ServeHTTP_wrongRequest() {
	r, _ := setupRouter()

	f1 := new(float64)
	*f1 = 1.25
	m := model.Metric{
		MType: "gauge", ID: "metric1", Value: f1,
	}
	b, _ := json.Marshal(m)
	req, _ := http.NewRequest(http.MethodPost, address+postMetrics, bytes.NewReader(b))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 400
	// {"error":"Bad request"}
}

func ExamplePingStorage_ServeHTTP_ok() {
	r, srv := setupRouter()
	srv.On("PingStorage").Once().Return(nil)
	req, _ := http.NewRequest(http.MethodGet, pingStorage, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	// Output:
	// 200
}

func ExamplePingStorage_ServeHTTP_serviceError() {
	r, srv := setupRouter()
	srv.On("PingStorage").Once().Return(errors.New("PingStorage error"))
	req, _ := http.NewRequest(http.MethodGet, pingStorage, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	// Output:
	// 500
}

func ExampleGetMetric_ServeHTTP_getGaugeMetric_ok() {
	r, srv := setupRouter()
	req, _ := http.NewRequest(http.MethodGet, address+getGaugePath, nil)
	rr := httptest.NewRecorder()

	f := new(float64)
	*f = 1.23
	m := &model.Metric{MType: "gauge", ID: "metric1", Value: f}
	srv.On("GetMetric", m.MType, m.ID).Once().Return(m, nil)

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 200
	// 1.23
}

func ExampleGetMetric_ServeHTTP_getCounterMetric_ok() {
	r, srv := setupRouter()
	req, _ := http.NewRequest(http.MethodGet, address+getCounterPath, nil)
	rr := httptest.NewRecorder()

	f := new(int64)
	*f = 1
	m := &model.Metric{MType: "counter", ID: "metric1", Delta: f}
	srv.On("GetMetric", m.MType, m.ID).Once().Return(m, nil)

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 200
	// 1
}

func ExampleGetMetric_ServeHTTP_notFound() {
	r, srv := setupRouter()
	req, _ := http.NewRequest(http.MethodGet, address+getGaugePath, nil)
	rr := httptest.NewRecorder()

	f := new(float64)
	*f = 1.23
	m := &model.Metric{MType: "gauge", ID: "metric1"}
	srv.On("GetMetric", m.MType, m.ID).Once().Return(nil, errors.New("err"))

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 404
	// {"error":"Not found"}
}

func ExampleGetMetrics_ServeHTTP_ok() {
	r, srv := setupRouter()
	req, _ := http.NewRequest(http.MethodGet, address+getAllPath, nil)
	rr := httptest.NewRecorder()

	f1, f2, i := new(float64), new(float64), new(int64)
	*f1, *f2, *i = 1.23, 1.24, 10

	m1 := model.Metric{MType: "counter", ID: "metric1", Delta: i}
	m2 := model.Metric{MType: "gauge", ID: "metric1", Value: f1}
	m3 := model.Metric{MType: "gauge", ID: "metric2", Value: f2}
	d := model.Data(map[string]map[string]model.Metric{
		"counter": {"metric1": m1},
		"gauge":   {"metric1": m2, "metric2": m3},
	})
	srv.On("GetMetrics").Once().Return(d, nil)

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	// Output:
	// 200
}

func ExampleGetMetrics_ServeHTTP_internalServerError() {
	r, srv := setupRouter()
	req, _ := http.NewRequest(http.MethodGet, address+getAllPath, nil)
	rr := httptest.NewRecorder()
	srv.On("GetMetrics").Once().Return(nil, errors.New("internal server error"))

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 500
	// {"error":"Internal server error"}
}

func ExampleGetMetricV2_ServeHTTP_ok() {
	r, srv := setupRouter()
	b := []byte(`{"id": "metric1", "type":"gauge"}`)
	req, _ := http.NewRequest(http.MethodPost, address+"/value/", bytes.NewReader(b))
	rr := httptest.NewRecorder()

	f := new(float64)
	*f = 1.25

	m := &model.Metric{MType: "gauge", ID: "metric1", Value: f}
	srv.On("GetMetric", m.MType, m.ID).Once().Return(m, nil)

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 200
	// {"id":"metric1","type":"gauge","value":1.25}
}

func ExampleGetMetricV2_ServeHTTP_invalidData() {
	r, _ := setupRouter()
	b := []byte(`{"id": "metric1", "type":"gauge}`)
	req, _ := http.NewRequest(http.MethodPost, address+"/value/", bytes.NewReader(b))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 400
	// {"error":"Bad request"}
}

func ExampleGetMetricV2_ServeHTTP_notFound() {
	r, srv := setupRouter()
	b := []byte(`{"id": "metric2", "type":"gauge"}`)
	req, _ := http.NewRequest(http.MethodPost, address+"/value/", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	srv.On("GetMetric", "gauge", "metric2").Once().Return(nil, errors.New("err"))
	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 404
	// {"error":"Not found"}
}

func ExamplePostMetricV2_ServeHTTP_ok() {
	r, srv := setupRouter()
	f := new(float64)
	*f = 1.25
	m := model.Metric{MType: "gauge", ID: "metric1", Value: f}
	b, _ := json.Marshal(m)
	req, _ := http.NewRequest(http.MethodPost, address+"/update/", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	srv.On("SaveMetric", m).Once().Return(nil)

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 200
	// {"id":"metric1","type":"gauge","value":1.25}
}

func ExamplePostMetricV2_ServeHTTP_invalidData() {
	r, _ := setupRouter()
	b := []byte(`{"id": "metric1", "type": "gauge", "value": "1.25"}`)
	req, _ := http.NewRequest(http.MethodPost, address+"/update/", bytes.NewReader(b))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 400
	// {"error":"Bad request"}
}

func ExamplePostMetricV2_ServeHTTP_internalServerError() {
	r, srv := setupRouter()
	b := []byte(`{"id": "metric1", "type": "gauge", "value": 1.25}`)
	req, _ := http.NewRequest(http.MethodPost, address+"/update/", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	srv.On("SaveMetric", mmock.Anything).Once().Return(errors.New("err"))
	r.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))
	// Output:
	// 500
	// {"error":"Internal server error"}
}
