package main

import (
	"context"
	"crypto/rsa"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/config"
	"github.com/v-starostin/go-metrics/internal/crypto"
)

var (
	BuildVersion string
	BuildData    string
	BuildCommit  string
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Info().Msgf("Build version: %s", getValue(BuildVersion))
	logger.Info().Msgf("Build data: %s", getValue(BuildData))
	logger.Info().Msgf("Build commit: %s", getValue(BuildCommit))

	cfg, err := config.NewAgent()
	if err != nil {
		logger.Error().Err(err).Msg("Configuration error")
		return
	}
	client := &http.Client{
		Timeout: time.Minute,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	var publicKey *rsa.PublicKey
	if cfg.CryptoKey != "" {
		publicKey, err = crypto.LoadPublicKey(cfg.CryptoKey)
		if err != nil {
			logger.Error().Err(err).Msg("Error to load public key")
			return
		}
	}

	a := agent.New(&logger, client, cfg.ServerAddress, cfg.Key, publicKey)

	logger.Info().
		Int("pollInterval", cfg.PollInterval).
		Int("reportInterval", cfg.ReportInterval).
		Msg("Started collecting metrics")

	go a.CollectRuntimeMetrics(ctx, time.Duration(cfg.PollInterval)*time.Second)
	go a.CollectGopsutilMetrics(ctx, time.Duration(cfg.PollInterval)*time.Second)

	metrics := a.PrepareMetrics(ctx, time.Duration(cfg.ReportInterval)*time.Second)

	for i := 0; i < cfg.RateLimit; i++ {
		go a.Retry(ctx, 3, func(ctx context.Context) error {
			return a.SendMetrics(ctx, metrics)
		}, 1*time.Second, 3*time.Second, 5*time.Second)
	}

	<-ctx.Done()
	if ctx.Err() != nil {
		logger.Info().Msgf("Received shutdown signal, stopping work: %v", ctx.Err())
	}

	logger.Info().Msg("Waiting 5 seconds to complete pending operations...")
	time.Sleep(5 * time.Second)

	logger.Info().Msg("Finished collecting metrics and shutting down gracefully.")
}

func getValue(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
