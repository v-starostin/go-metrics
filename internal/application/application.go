package application

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"net/http"
	_ "net/http/pprof"
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
	"github.com/v-starostin/go-metrics/internal/crypto"
	"github.com/v-starostin/go-metrics/internal/handler"
	"github.com/v-starostin/go-metrics/internal/repository"
	"github.com/v-starostin/go-metrics/internal/service"
)

var (
	BuildVersion string
	BuildData    string
	BuildCommit  string
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

func (s *Server) RegisterHandlers(
	ctx context.Context,
	srv handler.Service,
	key,
	trustedSubnet string,
	privateKey *rsa.PrivateKey,
) {
	getMetricHandler := handler.NewGetMetric(ctx, s.logger, srv, key)
	getMetricsHandler := handler.NewGetMetrics(ctx, s.logger, srv, key)
	getMetricV2Handler := handler.NewGetMetricV2(ctx, s.logger, srv, key)
	postMetricHandler := handler.NewPostMetric(ctx, s.logger, srv)
	postMetricV2Handler := handler.NewPostMetricV2(ctx, s.logger, srv)
	postMetrics := handler.NewPostMetrics(ctx, s.logger, srv, privateKey)
	pingStorage := handler.NewPingStorage(ctx, s.logger, srv)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&handler.LogFormatter{Logger: s.logger}))
		r.Use(handler.CheckIP(trustedSubnet))
		r.Use(handler.CheckHash(key))
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

	if err = db.Ping(); err != nil {
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
	logger.Info().Msgf("Build version: %s", getValue(BuildVersion))
	logger.Info().Msgf("Build data: %s", getValue(BuildData))
	logger.Info().Msgf("Build commit: %s", getValue(BuildCommit))

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
	var privateKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		privateKey, err = crypto.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			logger.Error().Err(err).Msg("Error to load private key")
			return
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	svc := service.New(&logger, repo)
	server := NewServer(&logger, cfg.ServerAddress)
	server.RegisterHandlers(ctx, svc, cfg.Key, cfg.TrustedSubnet, privateKey)

	f := handler.NewFile1(svc)

	if *cfg.Restore {
		if err := f.RestoreFromFile(); err != nil {
			logger.Error().Err(err).Msg("Failed to restore storage from file")
		}
		logger.Info().Msg("Storage has been restored from file")
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Shutdown server error")
		return
	}

	s.logger.Info().Msg("Server stopped gracefully")
}

func getValue(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
