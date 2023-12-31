package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/config"
	"github.com/v-starostin/go-metrics/internal/model"
)

func main() {
	metrics := make([]model.Metric, len(model.GaugeMetrics)+2)
	counter := int64(0)
	cfg := config.NewAgent()
	poll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	report := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	client := &http.Client{
		Timeout: time.Minute,
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log.Printf("Started gathering metrics with pollInterval: %v, reportInterval: %v", cfg.PollInterval, cfg.ReportInterval)
loop:
	for {
		select {
		case <-poll.C:
			agent.CollectMetrics(metrics, &counter)
			log.Printf("\ncollecting: %+v\n\n", metrics)
		case <-report.C:
			fmt.Printf("\nsending: %+v\n\n", metrics)
			if err := agent.SendMetrics(ctx, client, metrics, cfg.ServerAddress); err != nil {
				log.Fatal(err)
			}
		case <-ctx.Done():
			log.Println(ctx.Err())
			poll.Stop()
			report.Stop()
			break loop
		}
	}
	log.Println("Finished gathering metrics")
}
