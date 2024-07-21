package handler

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
)

// PingStorage is a struct that handles HTTP request for pinging the DB.
type PingStorage struct {
	ctx     context.Context
	logger  *zerolog.Logger
	service Service
}

// NewPingStorage creates a new handler.
func NewPingStorage(ctx context.Context, l *zerolog.Logger, srv Service) *PingStorage {
	return &PingStorage{
		ctx:     ctx,
		logger:  l,
		service: srv,
	}
}

// ServeHTTP handles HTTP requests for retrieving a specific metric.
func (h *PingStorage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.service.PingStorage(h.ctx); err != nil {
		h.logger.Error().Err(err).Msg("Pinging DB error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
