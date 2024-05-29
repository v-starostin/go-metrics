package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

// Service  defines methods for saving, retrieving, and managing metrics.
type Service interface {
	SaveMetric(m model.Metric) error
	SaveMetrics(m []model.Metric) error
	GetMetric(mtype, mname string) (*model.Metric, error)
	GetMetrics() (model.Data, error)
	PingStorage() error
	WriteToFile() error
	RestoreFromFile() error
}

// GetMetric is a struct that handles HTTP requests for retrieving metrics.
type GetMetric struct {
	logger  *zerolog.Logger
	service Service
	key     string
}

// NewGetMetric creates a new handler.
func NewGetMetric(l *zerolog.Logger, srv Service) *GetMetric {
	return &GetMetric{
		logger:  l,
		service: srv,
	}
}

// ServeHTTP handles HTTP requests for retrieving a specific metric.
func (h *GetMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	mname := chi.URLParam(r, "name")

	metric, err := h.service.GetMetric(mtype, mname)
	if err != nil {
		writeResponse(w, http.StatusNotFound, model.Error{Error: "Not found"})
		return
	}
	h.logger.Info().Any("metric", metric).Msg("Received metric from storage")

	if h.key != "" {
		w.Header().Add("HashSHA256", sign(metric, h.key))
	}

	switch mtype {
	case service.TypeGauge:
		writeResponse(w, http.StatusOK, *metric.Value)
	case service.TypeCounter:
		writeResponse(w, http.StatusOK, *metric.Delta)
	}
}
