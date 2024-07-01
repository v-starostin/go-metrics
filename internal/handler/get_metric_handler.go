package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

// Service  defines methods for saving, retrieving, and managing metrics.
type Service interface {
	SaveMetric(ctx context.Context, m model.Metric) error
	SaveMetrics(ctx context.Context, m []model.Metric) error
	GetMetric(ctx context.Context, mtype, mname string) (*model.Metric, error)
	GetMetrics(ctx context.Context) (model.Data, error)
	PingStorage(ctx context.Context) error
	WriteToFile() error
	RestoreFromFile() error
}

// GetMetric is a struct that handles HTTP requests for retrieving metrics.
type GetMetric struct {
	ctx     context.Context
	logger  *zerolog.Logger
	service Service
	key     string
}

// NewGetMetric creates a new handler.
func NewGetMetric(ctx context.Context, l *zerolog.Logger, srv Service, k string) *GetMetric {
	return &GetMetric{
		ctx:     ctx,
		logger:  l,
		service: srv,
		key:     k,
	}
}

// ServeHTTP handles HTTP requests for retrieving a specific metric.
func (h *GetMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	mname := chi.URLParam(r, "name")

	metric, err := h.service.GetMetric(h.ctx, mtype, mname)
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
