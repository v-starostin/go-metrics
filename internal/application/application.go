package application

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

type Server struct {
	srv    *http.Server
	logger *zerolog.Logger
}

func NewServer(l *zerolog.Logger, addr string) *Server {
	return &Server{
		srv:    &http.Server{Addr: addr},
		logger: l,
	}
}
func (s *Server) RegisterHandlers(srv handler.Service, key string) {
	getMetricHandler := handler.NewGetMetric(s.logger, srv)
	getMetricsHandler := handler.NewGetMetrics(s.logger, srv)
	getMetricV2Handler := handler.NewGetMetricV2(s.logger, srv)
	postMetricHandler := handler.NewPostMetric(s.logger, srv)
	postMetricV2Handler := handler.NewPostMetricV2(s.logger, srv)
	postMetrics := handler.NewPostMetrics(s.logger, srv)
	pingStorage := handler.NewPingStorage(s.logger, srv)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&handler.LogFormatter{Logger: s.logger}))
		r.Use(handler.CheckHash(s.logger, key))
		r.Use(middleware.Compress(5, "text/html", "application/json"))
		r.Use(handler.Decompress(s.logger))
		r.Use(middleware.Recoverer)
		r.Method(http.MethodPost, "/update/{type}/{name}/{value}", postMetricHandler)
		r.Method(http.MethodGet, "/value/{type}/{name}", getMetricHandler)
		r.Method(http.MethodGet, "/", getMetricsHandler)
		r.Method(http.MethodPost, "/updates/", postMetrics)
		r.Method(http.MethodPost, "/update/", postMetricV2Handler)
		r.Method(http.MethodPost, "/value/", getMetricV2Handler)
		r.Method(http.MethodGet, "/ping", pingStorage)
	})

	s.srv.Handler = r
}

func ConnectDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDNS)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	instance, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance("file://db", "postgres", instance)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return db, nil
}

func Run() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.NewServer()
	if err != nil {
		logger.Error().Err(err).Msg("Configuration error")
		return
	}

	var repo service.Repository
	var db *sql.DB
	if cfg.DatabaseDNS != "" {
		db, err = ConnectDB(&cfg)
		if err != nil {
			logger.Error().Err(err).Msg("DB initializing error")
			return
		}
		defer db.Close()
		repo = repository.NewStorage(&logger, db)
	} else {
		repo = repository.NewMemStorage(&logger, *cfg.StoreInterval, cfg.FileStoragePath)
	}

	svc := service.New(&logger, repo)
	server := NewServer(&logger, cfg.ServerAddress)
	server.RegisterHandlers(svc, cfg.Key)

	f := handler.NewFile1(svc)

	if *cfg.Restore {
		if err := f.RestoreFromFile(); err != nil {
			logger.Error().Err(err).Msg("Failed to restore storage from file")
		}
		logger.Info().Msg("Storage has been restored from file")
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
					if err := f.WriteToFile(); err != nil {
						logger.Error().Err(err).Msg("Failed to write storage content to file")
					}
				case <-ctx.Done():
					ticker.Stop()
					break loop
				}
			}
		}()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go server.ListenAndServe(&cfg)
	go server.HandleShutdown(ctx, wg, f, cfg.DatabaseDNS != "")

	wg.Wait()
}

func (s *Server) ListenAndServe(cfg *config.Config) {
	s.logger.Info().Msgf("Server is listerning on %s", cfg.ServerAddress)
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error().Err(err).Msg("Server error")
		return
	}
}

func (s *Server) HandleShutdown(ctx context.Context, wg *sync.WaitGroup, f *handler.File, dbEnabled bool) {
	defer wg.Done()

	<-ctx.Done()
	s.logger.Info().Msg("Shutdown signal received")

	if !dbEnabled {
		if err := f.WriteToFile(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to write storage content to file")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := s.srv.Shutdown(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Shutdown server error")
		return
	}

	s.logger.Info().Msg("Server stopped gracefully")
}
