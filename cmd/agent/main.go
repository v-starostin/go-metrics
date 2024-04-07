package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/config"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.NewAgent()
	if err != nil {
		logger.Error().Err(err).Msg("Configuration error")
		return
	}
	client := &http.Client{
		Timeout: time.Minute,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	a := agent.New(&logger, client, cfg.ServerAddress, cfg.Key)

	logger.Info().
		Int("pollInterval", cfg.PollInterval).
		Int("reportInterval", cfg.ReportInterval).
		Msg("Started collecting metrics")

	go a.CollectRuntimeMetrics(ctx, time.Duration(cfg.PollInterval)*time.Second)
	go a.CollectGopsutilMetrics(ctx, time.Duration(cfg.PollInterval)*time.Second)

	metrics := a.PrepareMetrics(ctx, time.Duration(cfg.ReportInterval)*time.Second)

	for i := 0; i < cfg.RateLimit; i++ {
		//go a.Retry(ctx, 3, func(ctx context.Context) error {
		//	return a.SendMetrics(ctx, metrics)
		go a.SendMetrics(ctx, metrics)
		//go a.PrintMetrics(ctx, metrics)
		//}, 1*time.Second, 3*time.Second, 5*time.Second)
	}

	<-ctx.Done()
	logger.Info().Msg("Finished collecting metrics")
}
