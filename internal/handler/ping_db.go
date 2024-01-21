package handler

import (
	"database/sql"
	"net/http"

	"github.com/rs/zerolog"
)

func DBPing(logger *zerolog.Logger, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			return
		}
		if err := db.Ping(); err != nil {
			logger.Error().Err(err).Msg("Pinging DB error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
