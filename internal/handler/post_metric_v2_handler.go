package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
)

// PostMetricV2 is a struct that handles HTTP request for posting the metrics.
type PostMetricV2 struct {
	logger  *zerolog.Logger
	service Service
}

// NewPostMetricV2 creates a new handler.
func NewPostMetricV2(l *zerolog.Logger, srv Service) *PostMetricV2 {
	return &PostMetricV2{
		logger:  l,
		service: srv,
	}
}

// ServeHTTP handles HTTP requests for retrieving a specific metric.
func (h *PostMetricV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Any("req", r.Body).Msg("Request body")

	var req model.Metric
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
		return
	}
	h.logger.Info().Any("req", req).Msg("Decoded request body")

	if err := h.service.SaveMetric(req); err != nil {
		h.logger.Error().Err(err).Msg("SaveMetric method error")
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}

	writeResponse(w, http.StatusOK, req)
}
