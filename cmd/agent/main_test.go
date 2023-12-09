package main_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/v-starostin/go-metrics/internal/agent"
	"github.com/v-starostin/go-metrics/internal/mock"
	"github.com/v-starostin/go-metrics/internal/model"
)

// Had to move it here from internal/agent since GHActions checks expect agent tests in cmd/agent
func TestSendMetrics(t *testing.T) {
	var ctx = context.Background()
	client := &mock.HTTPClient{}
	metrics := []model.Metric{
		{Type: "gauge", Name: "metric1", Value: 12},
		{Type: "counter", Name: "metric3", Value: 3},
	}
	httpServerAddress := "0.0.0.0:8080"
	t.Run("good case", func(t *testing.T) {
		{
			req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8080/update/gauge/metric1/12", nil)
			assert.NoError(t, err)
			req.Header.Add("Content-Type", "text/plain")
			res := &http.Response{}
			res.StatusCode = http.StatusOK
			res.Body = io.NopCloser(bytes.NewBufferString("response"))
			client.On("Do", req).Return(res, nil)
		}
		{
			req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8080/update/counter/metric3/3", nil)
			assert.NoError(t, err)
			req.Header.Add("Content-Type", "text/plain")
			res := &http.Response{}
			res.StatusCode = http.StatusOK
			res.Body = io.NopCloser(bytes.NewBufferString("response"))
			client.On("Do", req).Return(res, nil)
		}

		err := agent.SendMetrics(ctx, client, metrics, httpServerAddress)
		assert.NoError(t, err)
	})
}
