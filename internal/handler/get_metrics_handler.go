package handler

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
)

type GetMetrics struct {
	logger  *zerolog.Logger
	service Service
}

func NewGetMetrics(l *zerolog.Logger, srv Service) *GetMetrics {
	return &GetMetrics{
		logger:  l,
		service: srv,
	}
}

func (h *GetMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.GetMetrics()
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}
	h.logger.Info().Any("metrics", metrics).Msg("Received metrics from storage")

	tmpl, err := template.New("metrics").Parse(model.HTMLTemplateString)
	if err != nil {
		// to be fixed
		writeResponse(w, http.StatusOK, model.Error{Error: "Internal server error"})
		return
	}
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, metrics); err != nil {
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}
	writeResponse(w, http.StatusOK, buf.Bytes())
}
