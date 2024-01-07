package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

type PostMetric struct {
	logger  *zerolog.Logger
	service Service
}

func NewPostMetric(l *zerolog.Logger, srv Service) *PostMetric {
	return &PostMetric{
		logger:  l,
		service: srv,
	}
}

func (h *PostMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	mname := chi.URLParam(r, "name")
	mvalue := chi.URLParam(r, "value")

	if mtype != service.TypeCounter && mtype != service.TypeGauge {
		writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
		return
	}

	var m model.Metric

	switch mtype {
	case service.TypeCounter:
		delta, err := strconv.ParseInt(mvalue, 10, 0)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
			//return service.ErrParseMetric
			return
		}
		m = model.Metric{
			ID:    mname,
			MType: mtype,
			Delta: &delta,
		}
	case service.TypeGauge:
		value, err := strconv.ParseFloat(mvalue, 64)
		if err != nil {
			//return service.ErrParseMetric
			writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
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
			writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
			return
		}
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}

	writeResponse(w, http.StatusOK, fmt.Sprintf("metric %s of type %s with value %v has been set successfully", mname, mtype, mvalue))
}
