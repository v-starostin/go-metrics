package handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"html/template"
	"net/http"
	"strconv"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

type Handler struct {
	logger  *zerolog.Logger
	service Service
}

func New(l *zerolog.Logger, s Service) *Handler {
	return &Handler{
		logger:  l,
		service: s,
	}
}

type Service interface {
	//SaveMetric(mtype, mname, mvalue string) error

	SaveMetric(m model.Metric) error
	GetMetric(mtype, mname string) (*model.Metric, error)
	GetMetrics() (model.Data, error)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	mname := chi.URLParam(r, "name")
	mvalue := chi.URLParam(r, "value")

	switch r.Method {
	case http.MethodPost:
		saveMetric(mtype, mname, mvalue, w, h)
	case http.MethodGet:
		getMetrics(mtype, mname, w, h)
	}
}

func saveMetric(mtype, mname, mvalue string, w http.ResponseWriter, h Handler) {
	if mtype != service.TypeCounter && mtype != service.TypeGauge {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var m model.Metric
	switch mtype {
	case service.TypeCounter:
		value, err := strconv.ParseInt(mvalue, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "bad request", http.StatusBadRequest)
			//return service.ErrParseMetric
			return
		}
		m = model.Metric{
			ID:    mname,
			MType: mtype,
			Delta: &value,
		}

	case service.TypeGauge:
		value, err := strconv.ParseFloat(mvalue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "bad request", http.StatusBadRequest)
			//return service.ErrParseMetric
			return
		}
		m = model.Metric{
			ID:    mname,
			MType: mtype,
			Value: &value,
		}
	}

	if err := h.service.SaveMetric(m); err != nil {
		if errors.Is(err, service.ErrParseMetric) {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("metric %s of type %s with value %v has been set successfully", mname, mtype, mvalue)))
}

func getMetrics(mtype, mname string, w http.ResponseWriter, h Handler) {
	if mtype == "" || mname == "" {
		metrics, err := h.service.GetMetrics()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		h.logger.Info().Msgf("Recieved metrics from storage: %+v\n", metrics)

		tmpl, err := template.New("metrics").Parse(model.HTMLTemplateString)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, metrics); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
		return
	}

	metric, err := h.service.GetMetric(mtype, mname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}
	h.logger.Info().Msgf("Recieved metric from storage: %+v", metric)

	switch mtype {
	case service.TypeGauge:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatFloat(*metric.Value, 'f', -1, 64)))
	case service.TypeCounter:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatInt(*metric.Delta, 10)))
	}
}
