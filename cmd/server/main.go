package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/v-starostin/go-metrics/internal/application"
	"github.com/v-starostin/go-metrics/internal/config"
	"github.com/v-starostin/go-metrics/internal/crypto"
	"github.com/v-starostin/go-metrics/internal/handler"
	"github.com/v-starostin/go-metrics/internal/pb"
	"github.com/v-starostin/go-metrics/internal/repository"
	"github.com/v-starostin/go-metrics/internal/service"
)

const GRPCPort = 8081

var (
	BuildVersion string
	BuildData    string
	BuildCommit  string
)

func main() {
	//application.Run()

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
		db, err = application.ConnectDB(&cfg)
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
	h := handler.NewPostMetrics(ctx, &logger, svc, privateKey)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(
		grpcrecovery.UnaryServerInterceptor(),
	)))
	pb.RegisterGoMetricsServer(grpcServer, h)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", GRPCPort))
	if err != nil {
		logger.Error().Err(err).Msg("Listening gRPC error")
		return
	}

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

	go func() {
		logger.Info().Msgf("GRPC server is listening on :%d", GRPCPort)
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error().Err(err).Msg("Server error")
			return
		}
	}()

	go handleShutdown(ctx, &logger, grpcServer, wg, f, cfg.DatabaseDNS != "")

	wg.Wait()
}

func handleShutdown(ctx context.Context, l *zerolog.Logger, grpcServer *grpc.Server, wg *sync.WaitGroup, f *handler.File, dbEnabled bool) {
	defer wg.Done()

	<-ctx.Done()
	l.Info().Msg("Shutdown signal received")

	if !dbEnabled {
		if err := f.WriteToFile(); err != nil {
			l.Error().Err(err).Msg("Failed to write storage content to file")
		}
	}

	grpcServer.GracefulStop()
	l.Info().Msg("Server stopped gracefully")
}

func getValue(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
