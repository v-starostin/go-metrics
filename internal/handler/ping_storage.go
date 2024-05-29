package handler

import (
	"net/http"

	"github.com/rs/zerolog"
)

// PingStorage is a struct that handles HTTP request for pinging the DB.
type PingStorage struct {
	logger  *zerolog.Logger
	service Service
}

// NewPingStorage creates a new handler.
func NewPingStorage(l *zerolog.Logger, srv Service) *PingStorage {
	return &PingStorage{
		logger:  l,
		service: srv,
	}
}

// ServeHTTP handles HTTP requests for retrieving a specific metric.
func (h *PingStorage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.service.PingStorage(); err != nil {
		h.logger.Error().Err(err).Msg("Pinging DB error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
