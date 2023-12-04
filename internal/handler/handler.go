package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/v-starostin/go-metrics/internal/service"
)

type Handler struct {
	service Service
}

func New(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

type Service interface {
	Save(mtype, mname, mvalue string) error
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.EscapedPath()
	trimmed := strings.Trim(p, "/")
	url := strings.Split(trimmed, "/")

	if url == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(url) != 4 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	cmd, mtype, mname, mvalue := url[0], url[1], url[2], url[3]

	if r.Method == http.MethodPost {
		if cmd != service.CmdUpdate || mtype != service.TypeCounter && mtype != service.TypeGauge {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := h.service.Save(mtype, mname, mvalue); err != nil {
			if errors.Is(err, service.ErrParseMetric) {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}

			http.Error(w, "service error", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "text-plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("metric %s of type %s with value %v has been set successfully", mname, mtype, mvalue)))
	} else {
		http.Error(w, fmt.Sprintf("method %s is not supported", r.Method), http.StatusBadRequest)
		return
	}

}
