package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
)

// PostMetrics is a struct that handles HTTP request for posting the metrics.
type PostMetrics struct {
	logger  *zerolog.Logger
	service Service
}

// NewPostMetrics creates a new handler.
func NewPostMetrics(l *zerolog.Logger, srv Service) *PostMetrics {
	return &PostMetrics{
		logger:  l,
		service: srv,
	}
}

// ServeHTTP handles HTTP requests for retrieving a specific metric.
func (h *PostMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req []model.Metric
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
		return
	}
	h.logger.Info().Any("req", req).Msg("Decoded request body")

	if err := h.service.SaveMetrics(req); err != nil {
		h.logger.Error().Err(err).Msg("SaveMetric method error")
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}

	writeResponse(w, http.StatusOK, req)
}
