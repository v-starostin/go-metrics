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
		logger.Fatal().Err(err).Msg("Configuration error")
	}
	poll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	report := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	client := &http.Client{
		Timeout: time.Minute,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	a := agent.New(&logger, client, cfg.ServerAddress)

	logger.Info().
		Int("pollInterval", cfg.PollInterval).
		Int("reportInterval", cfg.ReportInterval).
		Msg("Started collecting metrics")

loop:
	for {
		select {
		case <-poll.C:
			a.CollectMetrics()
			logger.Info().Interface("metrics", a.Metrics).Msg("Metrics collected")
		case <-report.C:
			a.SendMetrics(ctx)
			logger.Info().Interface("metrics", a.Metrics).Msg("Metrics sent")
		case <-ctx.Done():
			logger.Info().Err(ctx.Err()).Send()
			poll.Stop()
			report.Stop()
			break loop
		}
	}
	logger.Info().Msg("Finished collecting metrics")
}
