package main_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	mmock "github.com/stretchr/testify/mock"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/mock"
	"github.com/v-starostin/go-metrics/internal/model"
)

// Had to move it here from internal/agent since GHActions checks expect agent tests in cmd/agent
func TestSendMetrics(t *testing.T) {
	ctx := context.Background()
	client := &mock.GoMetricsClient{}
	metrics := []model.AgentMetric{
		{MType: "gauge", ID: "metric1", Value: float64(10)},
	}

	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key", nil)

	t.Run("good case", func(t *testing.T) {
		ch := make(chan []model.AgentMetric)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch <- metrics
		}()
		go func() {
			wg.Wait()
			close(ch)
		}()
		res := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("test")),
		}
		client.On("Do", mmock.Anything).Once().Return(res, nil)
		err := a.SendMetrics(ctx, ch, "192.168.8.22")
		assert.NoError(t, err)
	})

	t.Run("bad case", func(t *testing.T) {
		ch := make(chan []model.AgentMetric)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch <- metrics
		}()
		go func() {
			wg.Wait()
			close(ch)
		}()

		client.On("Do", mmock.Anything).Once().Return(nil, fmt.Errorf("err"))
		err := a.SendMetrics(ctx, ch, "192.168.8.22")
		assert.EqualError(t, err, "err")
	})
}

func TestRetry(t *testing.T) {
	ctx := context.Background()
	client := &mock.GoMetricsClient{}

	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key", nil)

	t.Run("good case", func(t *testing.T) {
		err := a.Retry(ctx, 3, func(ctx context.Context) error {
			return nil
		}, 10*time.Millisecond, 20*time.Millisecond, 30*time.Millisecond)

		assert.NoError(t, err)
	})

	t.Run("good case - 3th try is successful", func(t *testing.T) {
		var counter int

		err := a.Retry(ctx, 3, func(ctx context.Context) error {
			if counter == 3 {
				return nil
			}
			counter++
			return fmt.Errorf("err")
		}, 10*time.Millisecond, 20*time.Millisecond, 30*time.Millisecond)

		assert.NoError(t, err)
	})

	t.Run("bad case - no success after 3 tries", func(t *testing.T) {
		err := a.Retry(ctx, 3, func(ctx context.Context) error {
			return fmt.Errorf("err")
		}, 10*time.Millisecond, 20*time.Millisecond, 30*time.Millisecond)

		assert.EqualError(t, err, "err")
	})

	t.Run("bad case - context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()

		err := a.Retry(ctx, 3, func(ctx context.Context) error {
			return fmt.Errorf("err")
		}, 10*time.Millisecond, 20*time.Millisecond, 30*time.Millisecond)

		assert.EqualError(t, err, "context deadline exceeded")
	})

	t.Run("bad case - context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		err := a.Retry(ctx, 3, func(ctx context.Context) error {
			cancel()
			return fmt.Errorf("err")
		}, 10*time.Millisecond, 20*time.Millisecond, 30*time.Millisecond)

		assert.EqualError(t, err, "context canceled")
	})
}

func TestCollectGopsutilMetrics(t *testing.T) {
	ctx := context.Background()
	client := &mock.GoMetricsClient{}

	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key", nil)

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.CollectGopsutilMetrics(ctx, 15*time.Millisecond)
		}()
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		wg.Wait()
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.CollectGopsutilMetrics(ctx, 15*time.Millisecond)
		}()
		wg.Wait()
	})
}

func TestCollectRuntimeMetrics(t *testing.T) {
	ctx := context.Background()
	client := &mock.GoMetricsClient{}

	a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "key", nil)

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.CollectRuntimeMetrics(ctx, 15*time.Millisecond)
		}()
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		wg.Wait()
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.CollectRuntimeMetrics(ctx, 15*time.Millisecond)
		}()
		wg.Wait()
	})
}
