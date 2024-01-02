package main_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	mmock "github.com/stretchr/testify/mock"
	"github.com/v-starostin/go-metrics/internal/mock"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/model"
)

// Had to move it here from internal/agent since GHActions checks expect agent tests in cmd/agent
func TestSendMetrics(t *testing.T) {
	var ctx = context.Background()
	client := &mock.HTTPClient{}
	metrics := []model.AgentMetric{
		{MType: "gauge", ID: "metric1", Value: float64(10)},
	}
	t.Run("good case", func(t *testing.T) {
		res := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("test")),
		}
		client.On("Do", mmock.Anything).Once().Return(res, nil)
		err := agent.SendMetrics(ctx, &zerolog.Logger{}, client, metrics, "0.0.0.0:8080")
		assert.NoError(t, err)
	})
}
