package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

type Service interface {
	SaveMetric(m model.Metric) error
	SaveMetrics(m []model.Metric) error
	GetMetric(mtype, mname string) (*model.Metric, error)
	GetMetrics() (model.Data, error)
	PingStorage() error
	WriteToFile() error
	RestoreFromFile() error
}

type GetMetric struct {
	logger  *zerolog.Logger
	service Service
	key     string
}

func NewGetMetric(l *zerolog.Logger, srv Service) *GetMetric {
	return &GetMetric{
		logger:  l,
		service: srv,
	}
}

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
		b, err := json.Marshal(metric)
		if err != nil {
			return
		}

		h1 := hmac.New(sha256.New, []byte(h.key))
		h1.Write(b)
		d := h1.Sum(nil)

		w.Header().Add("HashSHA256", hex.EncodeToString(d))
	}

	switch mtype {
	case service.TypeGauge:
		writeResponse(w, http.StatusOK, *metric.Value)
	case service.TypeCounter:
		writeResponse(w, http.StatusOK, *metric.Delta)
	}
}
