package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/config"
	"github.com/v-starostin/go-metrics/internal/model"
)

func main() {
	var metrics []model.Metric
	counter := int64(0)
	cfg := config.NewAgent()
	poll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	report := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	client := &http.Client{
		Timeout: time.Minute,
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

loop:
	for {
		log.Println("Agent is started")

		select {
		case <-poll.C:
			metrics = agent.CollectMetrics()
			counter++
			metrics = append(metrics, model.Metric{Type: "gauge", Name: "RandomValue", Value: rand.Float64()})
			log.Printf("collecting: %+v\n\n", metrics)
		case <-report.C:
			metrics = append(metrics, model.Metric{Type: "counter", Name: "PollCount", Value: counter})
			fmt.Printf("sending: %+v\n\n", metrics)
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
}
