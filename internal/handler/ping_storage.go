package handler

import (
	"net/http"

	"github.com/rs/zerolog"
)

type PingStorage struct {
	logger  *zerolog.Logger
	service Service
}

func NewPingStorage(l *zerolog.Logger, srv Service) *PingStorage {
	return &PingStorage{
		logger:  l,
		service: srv,
	}
}

func (h *PingStorage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.service.PingStorage(); err != nil {
		h.logger.Error().Err(err).Msg("Pinging DB error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
