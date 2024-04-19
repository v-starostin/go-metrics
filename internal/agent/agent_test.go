package agent_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/mock"
)

func BenchmarkCollectRuntimeMetrics(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	client := &mock.HTTPClient{}
	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.CollectRuntimeMetrics(ctx, 50*time.Millisecond)
	}
}

func BenchmarkCollectGopsutilMetrics(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	client := &mock.HTTPClient{}
	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.CollectGopsutilMetrics(ctx, 50*time.Millisecond)
	}
}
