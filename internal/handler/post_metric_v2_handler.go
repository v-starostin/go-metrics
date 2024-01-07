package handler

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/v-starostin/go-metrics/internal/model"
	"net/http"
)

type PostMetricV2 struct {
	logger  *zerolog.Logger
	service Service
}

func NewPostMetricV2(l *zerolog.Logger, srv Service) *PostMetricV2 {
	return &PostMetricV2{
		logger:  l,
		service: srv,
	}
}

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
