package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/config"
	"github.com/v-starostin/go-metrics/internal/handler"
	"github.com/v-starostin/go-metrics/internal/repository"
	"github.com/v-starostin/go-metrics/internal/service"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.NewServer()
	if err != nil {
		logger.Fatal().Err(err).Msg("Configuration error")
	}
	repo := repository.New(&logger)
	srv := service.New(&logger, repo)
	getMetricHandler := handler.NewGetMetric(&logger, srv)
	getMetricsHandler := handler.NewGetMetrics(&logger, srv)
	getMetricV2Handler := handler.NewGetMetricV2(&logger, srv)
	postMetricHandler := handler.NewPostMetric(&logger, srv)
	postMetricV2Handler := handler.NewPostMetricV2(&logger, srv)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&handler.LogFormatter{Logger: &logger}))
		r.Use(middleware.Compress(5, "text/html", "application/json"))
		r.Use(handler.Decompress(&logger))
		r.Use(middleware.Recoverer)
		r.Method(http.MethodPost, "/update/{type}/{name}/{value}", postMetricHandler)
		r.Method(http.MethodGet, "/value/{type}/{name}", getMetricHandler)
		r.Method(http.MethodGet, "/", getMetricsHandler)
		r.Method(http.MethodPost, "/update/", postMetricV2Handler)
		r.Method(http.MethodPost, "/value/", getMetricV2Handler)
	})

	logger.Info().Msgf("Server is listerning on %s", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal().Err(err).Msg("Server error")
	}
}
