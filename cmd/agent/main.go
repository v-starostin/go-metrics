package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/signal"
	"syscall"
	"time"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/model"
)

var (
	pollInterval   = 5 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var metrics []model.Metric
	var counter int64
	var poll = time.NewTicker(pollInterval)
	var report = time.NewTicker(reportInterval)

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
			if err := agent.SendMetrics(ctx, metrics); err != nil {
				log.Fatal(err)
			}
		case <-ctx.Done():
			log.Println(ctx.Err())
			poll.Stop()
			report.Stop()
			break loop
		}
	}

	log.Println("Agent is shutdowned gracefully")
}
