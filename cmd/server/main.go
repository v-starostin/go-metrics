package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	repo := repository.New(&logger, *cfg.StoreInterval, cfg.FileStoragePath)
	srv := service.New(&logger, repo)
	getMetricHandler := handler.NewGetMetric(&logger, srv)
	getMetricsHandler := handler.NewGetMetrics(&logger, srv)
	getMetricV2Handler := handler.NewGetMetricV2(&logger, srv)
	postMetricHandler := handler.NewPostMetric(&logger, srv)
	postMetricV2Handler := handler.NewPostMetricV2(&logger, srv)

	if *cfg.Restore {
		err := repo.RestoreFromFile()
		if err != nil {
			logger.Error().Err(err).Msg("Failed restore file")
		}
		logger.Info().Msg("Storage file has been restored")
	}

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

	server := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	ch := make(chan struct{})
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if *cfg.StoreInterval > 0 {
		ticker := time.NewTicker(time.Duration(*cfg.StoreInterval) * time.Second)

		go func() {
		loop:
			for {
				select {
				case <-ticker.C:
					if err := repo.WriteToFile(); err != nil {
						logger.Error().Err(err).Msg("Failed to write storage content to file")
					}
				case <-ctx.Done():
					ticker.Stop()
					if err := repo.WriteToFile(); err != nil {
						logger.Error().Err(err).Msg("Failed to write storage content to file")
					}
					ch <- struct{}{}
					break loop
				}
			}
		}()
	} else {
		go func() {
			<-ctx.Done()
			if err := repo.WriteToFile(); err != nil {
				logger.Error().Err(err).Msg("Failed to write storage content to file")
			}
			ch <- struct{}{}
		}()
	}

	go func() {
		logger.Info().Msgf("Server is listerning on %s", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	<-ch
	logger.Info().Msg("Shutdown signal received")

	cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(cctx); err != nil {
		logger.Fatal().Err(err).Msg("Shutdown server error")
	}

	logger.Info().Msg("Server stopped gracefully")
}
