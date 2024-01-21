package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	var db *sql.DB
	if cfg.DatabaseDNS != "" {
		db, err = sql.Open("pgx", cfg.DatabaseDNS)
		if err != nil {
			logger.Fatal().Err(err).Msg("DB initializing error")
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			logger.Fatal().Err(err).Msg("DB pinging error")
		}

		instance, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			log.Println(err)
			return
		}
		m, err := migrate.NewWithDatabaseInstance("file://db", "postgres", instance)
		if err != nil {
			log.Println(err)
			return
		}
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return
		}
	}

	var repo service.Repository
	if cfg.DatabaseDNS != "" {
		repo = repository.NewDB(db)
	} else {
		repo = repository.NewMemStorage(&logger, *cfg.StoreInterval, cfg.FileStoragePath)
	}
	srv := service.New(&logger, repo)
	getMetricHandler := handler.NewGetMetric(&logger, srv)
	getMetricsHandler := handler.NewGetMetrics(&logger, srv)
	getMetricV2Handler := handler.NewGetMetricV2(&logger, srv)
	postMetricHandler := handler.NewPostMetric(&logger, srv)
	postMetricV2Handler := handler.NewPostMetricV2(&logger, srv)
	postMetrics := handler.NewPostMetrics(&logger, srv)
	pingDB := handler.DBPing(&logger, db)

	if *cfg.Restore {
		err := repo.RestoreFromFile()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to restore storage from file")
		}
		logger.Info().Msg("Storage has been restored from file")
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
		r.Method(http.MethodPost, "/updates/", postMetrics)
		r.Method(http.MethodPost, "/update/", postMetricV2Handler)
		r.Method(http.MethodPost, "/value/", getMetricV2Handler)
		r.Method(http.MethodGet, "/ping", pingDB)
	})

	server := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

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
					break loop
				}
			}
		}()
	}

	go func() {
		logger.Info().Msgf("Server is listerning on %s", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	<-ctx.Done()
	logger.Info().Msg("Shutdown signal received")

	if err := repo.WriteToFile(); err != nil {
		logger.Error().Err(err).Msg("Failed to write storage content to file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Shutdown server error")
	}

	logger.Info().Msg("Server stopped gracefully")
}
