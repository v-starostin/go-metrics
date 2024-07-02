package handler

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/crypto"
	"github.com/v-starostin/go-metrics/internal/model"
)

// PostMetrics is a struct that handles HTTP request for posting the metrics.
type PostMetrics struct {
	ctx        context.Context
	logger     *zerolog.Logger
	service    Service
	privateKey *rsa.PrivateKey
}

// NewPostMetrics creates a new handler.
func NewPostMetrics(ctx context.Context, l *zerolog.Logger, srv Service, pk *rsa.PrivateKey) *PostMetrics {
	return &PostMetrics{
		ctx:        ctx,
		logger:     l,
		service:    srv,
		privateKey: pk,
	}
}

// ServeHTTP handles HTTP requests for retrieving a specific metric.
func (h *PostMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
	}

	var data []byte
	if h.privateKey != nil {
		data, err = crypto.RSADecrypt(h.privateKey, b)
		if err != nil {
			h.logger.Error().Err(err).Msg("Invalid incoming data")
			writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
		}
	}

	var req []model.Metric
	if err := json.Unmarshal(data, &req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, model.Error{Error: "Bad request"})
		return
	}
	h.logger.Info().Any("req", req).Msg("Decoded request body")

	if err := h.service.SaveMetrics(h.ctx, req); err != nil {
		h.logger.Error().Err(err).Msg("SaveMetric method error")
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}

	writeResponse(w, http.StatusOK, req)
}
