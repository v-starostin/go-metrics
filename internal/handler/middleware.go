package handler

import (
	"compress/gzip"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type LogFormatter struct {
	*zerolog.Logger
}

func (l *LogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	logger := l.With().
		Str("URI", r.RequestURI).
		Str("method", r.Method).
		Logger()

	return &LogEntry{&logger}
}

type LogEntry struct {
	*zerolog.Logger
}

func (l *LogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ any) {
	l.Info().
		Str("elapsed", elapsed.String()).
		Int("bytes", bytes).
		Int("status", status).
		Msg("Request handled")
}

func (l *LogEntry) Panic(v any, stack []byte) {
	l.Info().
		Interface("panic", v).
		Bytes("stack", stack).
		Msg("Panic handled")
}

func Decompress(l *zerolog.Logger) func(next http.Handler) http.Handler {
	//pool := sync.Pool{
	//	New: func() any { return new(gzip.Reader) },
	//}
	gr, err := gzip.NewReader(nil)
	if err != nil {
		l.Error().Msg("Error to create gzip reader")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encodingHeaders := r.Header.Values("Content-Encoding")
			if !slices.Contains(encodingHeaders, "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			//gr, ok := pool.Get().(*gzip.Reader)
			//if !ok {
			//	l.Error().Msg("Error to get Reader")
			//}
			//defer pool.Put(gr)

			if err := gr.Reset(r.Body); err != nil {
				l.Error().Err(err).Msg("Reset gzip reader error")
			}
			defer gr.Close()

			r.Body = gr
			next.ServeHTTP(w, r)
		})
	}
}
