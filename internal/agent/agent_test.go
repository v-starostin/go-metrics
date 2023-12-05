package agent_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/mock"
	"github.com/v-starostin/go-metrics/internal/model"
)

func TestSendMetrics(t *testing.T) {
	var ctx = context.Background()
	client := &mock.HTTPClient{}
	metrics := []model.Metric{
		{Type: "gauge", Name: "metric1", Value: 12},
		{Type: "counter", Name: "metric3", Value: 3},
	}
	t.Run("good case", func(t *testing.T) {
		client.On("Do", nil).Once().Return(nil, nil)
		err := agent.SendMetrics(ctx, client, metrics)
		assert.NoError(t, err)
	})
}
