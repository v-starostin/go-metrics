package handler

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

type Handler struct {
	service Service
}

func New(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

type Service interface {
	Save(mtype, mname, mvalue string) error
	Metric(mtype, mname string) (*model.Metric, error)
	Metrics() (model.Data, error)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	mname := chi.URLParam(r, "name")
	mvalue := chi.URLParam(r, "value")

	if r.Method == http.MethodPost {
		if mtype != service.TypeCounter && mtype != service.TypeGauge {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := h.service.Save(mtype, mname, mvalue); err != nil {
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

	if r.Method == http.MethodGet {
		if mtype == "" || mname == "" {
			metrics, err := h.service.Metrics()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			fmt.Printf("metrics: %+v\n", metrics)

			tmpl, err := template.New("metrics").Parse(model.HTMLTemplateString)
			if err != nil {
				return // ?
			}
			buf := bytes.Buffer{}
			if err := tmpl.Execute(&buf, metrics); err != nil {
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
			return
		}

		metric, err := h.service.Metric(mtype, mname)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		}

		if mtype == service.TypeGauge {
			mv := metric.Value.(float64)
			w.Write([]byte(strconv.FormatFloat(mv, 'f', -1, 64)))
		}
		if mtype == service.TypeCounter {
			mv := metric.Value.(int64)
			w.Write([]byte(strconv.FormatInt(mv, 10)))
		}
		w.WriteHeader(http.StatusOK)

		return
	}

}
