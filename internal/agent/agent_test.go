package agent_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	mmock "github.com/stretchr/testify/mock"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/mock"
	"github.com/v-starostin/go-metrics/internal/model"
)

func BenchmarkCollectRuntimeMetrics(b *testing.B) {
	client := &mock.GoMetricsClient{}
	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		b.StartTimer()
		a.CollectRuntimeMetrics(ctx, 10*time.Millisecond)
	}
}

func BenchmarkCollectGopsutilMetrics(b *testing.B) {
	client := &mock.GoMetricsClient{}
	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		b.StartTimer()
		a.CollectGopsutilMetrics(ctx, 10*time.Millisecond)
	}
}

func BenchmarkSendMetrics(b *testing.B) {
	ctx := context.Background()
	client := &mock.GoMetricsClient{}
	metrics := []model.AgentMetric{
		{MType: "gauge", ID: "metric0", Value: float64(10)},
		{MType: "gauge", ID: "metric1", Value: float64(11)},
		{MType: "gauge", ID: "metric2", Value: float64(12)},
		{MType: "gauge", ID: "metric3", Value: float64(13)},
		{MType: "gauge", ID: "metric4", Value: float64(14)},
		{MType: "gauge", ID: "metric5", Value: float64(15)},
		{MType: "gauge", ID: "metric6", Value: float64(16)},
		{MType: "gauge", ID: "metric7", Value: float64(17)},
		{MType: "gauge", ID: "metric8", Value: float64(18)},
		{MType: "gauge", ID: "metric9", Value: float64(19)},
	}

	var mm = make([][]model.AgentMetric, 0, 1000)
	for i := 0; i < 1000; i++ {
		mm = append(mm, metrics)
	}

	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key", nil)

	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("test")),
	}
	client.On("Do", mmock.Anything).Return(res, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		ch := make(chan []model.AgentMetric)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, m := range mm {
				ch <- m
			}
		}()
		go func() {
			wg.Wait()
			close(ch)
		}()

		b.StartTimer()
		a.SendMetrics(ctx, ch, "192.168.8.22")
	}
}
