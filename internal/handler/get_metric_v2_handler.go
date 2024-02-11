package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
)

type GetMetricV2 struct {
	logger  *zerolog.Logger
	service Service
	key     string
}

func NewGetMetricV2(l *zerolog.Logger, s Service) *GetMetricV2 {
	return &GetMetricV2{
		logger:  l,
		service: s,
	}
}

func (h *GetMetricV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Any("req", r.Body).Msg("Request body")

	var req model.Metric
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
		return
	}
	h.logger.Info().Any("req", req).Msg("Decoded request body")

	res, err := h.service.GetMetric(req.MType, req.ID)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetMetric method error")
		writeResponse(w, http.StatusNotFound, model.Error{Error: "Not found"})
		return
	}

	if h.key != "" {
		w.Header().Add("HashSHA256", sign(res, h.key))
	}

	writeResponse(w, http.StatusOK, res)
}
