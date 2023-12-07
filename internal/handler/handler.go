package handler

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"

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
	// cmd := chi.URLParam(r, "cmd")
	mtype := chi.URLParam(r, "type")
	mname := chi.URLParam(r, "name")
	mvalue := chi.URLParam(r, "value")

	// fmt.Println(cmd)
	fmt.Println(mtype)
	fmt.Println(mname)
	fmt.Println(mvalue)
	fmt.Println()

	if r.Method == http.MethodPost {
		if mtype != service.TypeCounter && mtype != service.TypeGauge {
			writeResponse(w, http.StatusBadRequest, model.Error{Error: "bad request"})
			return
		}

		if err := h.service.Save(mtype, mname, mvalue); err != nil {
			if errors.Is(err, service.ErrParseMetric) {
				writeResponse(w, http.StatusBadRequest, model.Error{Error: "bad request"})
				return
			}

			writeResponse(w, http.StatusInternalServerError, model.Error{Error: "internal server error"})
			return
		}
		writeResponse(w, http.StatusOK, fmt.Sprintf("metric %s of type %s with value %v has been set successfully", mname, mtype, mvalue))
	}

	if r.Method == http.MethodGet {
		if mtype == "" || mname == "" {
			metrics, err := h.service.Metrics()
			if err != nil {
				return
			}
			fmt.Printf("metrics: %+v\n", metrics)

			tmpl, err := template.New("metrics").Parse(model.HTMLTemplateString)
			if err != nil {
				return
			}
			buf := bytes.Buffer{}
			// fmt.Printf("buf: %+v\n", buf)
			if err := tmpl.Execute(&buf, metrics); err != nil {
				return
			}
			fmt.Printf("buf: %+v\n", buf.String())
			w.Write(buf.Bytes())
			// writeResponse(w, http.StatusOK, buf)
			return
		}

		metric, err := h.service.Metric(mtype, mname)
		if err != nil {
			writeResponse(w, http.StatusNotFound, model.Error{Error: "metric not found"})
			return
		}

		writeResponse(w, http.StatusOK, metric)

		return
	}

}
