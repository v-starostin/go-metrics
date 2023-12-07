package handler

import (
	"encoding/json"
	"net/http"
)

func writeResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "plain/text")
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	}

	w.WriteHeader(status)
	w.Write(b)
}
