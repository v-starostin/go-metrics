package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
)

type GetMetrics struct {
	logger  *zerolog.Logger
	service Service
	key     string
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
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, metrics); err != nil {
		writeResponse(w, http.StatusInternalServerError, model.Error{Error: "Internal server error"})
		return
	}

	if h.key != "" {
		b, err := json.Marshal(metrics)
		if err != nil {
			return
		}

		h1 := hmac.New(sha256.New, []byte(h.key))
		h1.Write(b)
		d := h1.Sum(nil)

		w.Header().Add("HashSHA256", hex.EncodeToString(d))
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
