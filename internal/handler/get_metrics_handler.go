package handler

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
)

// GetMetrics is a struct that handles HTTP requests for retrieving metrics.
type GetMetrics struct {
	logger  *zerolog.Logger
	service Service
	key     string
}

// NewGetMetrics creates a new handler.
func NewGetMetrics(l *zerolog.Logger, srv Service) *GetMetrics {
	return &GetMetrics{
		logger:  l,
		service: srv,
	}
}

// ServeHTTP handles HTTP requests for retrieving a specific metric.
func (h *GetMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.GetMetrics()
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}
	h.logger.Info().Any("metrics", metrics).Msg("Received metrics from storage")

	tmpl, err := template.New("metrics").Parse(model.HTMLTemplateString)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, metrics); err != nil {
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}

	if h.key != "" {
		w.Header().Add("HashSHA256", sign(metrics, h.key))
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
