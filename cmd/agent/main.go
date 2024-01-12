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
	"github.com/v-starostin/go-metrics/internal/model"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	metrics := make([]model.AgentMetric, len(model.GaugeMetrics)+2)
	//counter := int64(0)
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

	//pool := &sync.Pool{
	//	New: func() any { return gzip.NewWriter(io.Discard) },
	//}

	logger.Info().
		Int("pollInterval", cfg.PollInterval).
		Int("reportInterval", cfg.ReportInterval).
		Msg("Started collecting metrics")

	a := agent.New(&logger, client, cfg.ServerAddress)

loop:
	for {
		select {
		case <-poll.C:
			//agent.CollectMetrics(metrics, &counter)
			a.CollectMetrics1(metrics)
			logger.Info().Interface("metrics", metrics).Msg("Metrics collected")
		case <-report.C:
			//if err := agent.SendMetrics(ctx, &logger, client, metrics, cfg.ServerAddress, pool, ); err != nil {
			//	logger.Fatal().Err(err).Msg("Send metrics error")
			//}
			a.SendMetrics1(ctx, metrics)
			logger.Info().Interface("metrics", metrics).Msg("Metrics sent")
		case <-ctx.Done():
			logger.Info().Err(ctx.Err()).Send()
			poll.Stop()
			report.Stop()
			break loop
		}
	}
	logger.Info().Msg("Finished collecting metrics")
}
