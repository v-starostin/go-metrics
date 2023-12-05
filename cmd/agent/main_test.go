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

func TestSendMetrics(t *testing.T) {
	var ctx = context.Background()
	client := &mock.HTTPClient{}
	metrics := []model.Metric{
		{Type: "gauge", Name: "metric1", Value: 12},
		// {Type: "counter", Name: "metric3", Value: 3},
	}
	t.Run("good case", func(t *testing.T) {
		req := &http.Request{Method: http.MethodPost}
		req, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8080/update/gauge/metric1/12", nil)
		assert.NoError(t, err)
		req.Header.Add("Content-Type", "text/plain")

		resBody := "response"
		res := &http.Response{}
		res.StatusCode = http.StatusOK
		res.Body = io.NopCloser(bytes.NewBufferString(resBody))

		client.On("Do", req).Once().Return(res, nil)
		err = agent.SendMetrics(ctx, client, metrics)
		assert.NoError(t, err)
	})
}
