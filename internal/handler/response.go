package handler

import (
	"encoding/json"
	"net/http"
)

func writeResponse(w http.ResponseWriter, code int, v any) {
	w.Header().Add("Content-Type", "application/json")
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
		return
	}
	w.WriteHeader(code)
	w.Write(b)
}
