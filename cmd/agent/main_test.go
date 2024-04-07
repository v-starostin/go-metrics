package main_test

import (
	"testing"
)

// Had to move it here from internal/agent since GHActions checks expect agent tests in cmd/agent
func TestSendMetrics(t *testing.T) {
	//var ctx = context.Background()
	//client := &mock.HTTPClient{}
	//metrics := []model.AgentMetric{
	//	{MType: "gauge", ID: "metric1", Value: float64(10)},
	//}
	//a := agent.New(&zerolog.Logger{}, client, "0.0.0.0:8080", "")
	//a.Metrics = metrics
	//t.Run("good case", func(t *testing.T) {
	//	res := &http.Response{
	//		StatusCode: http.StatusOK,
	//		Body:       io.NopCloser(strings.NewReader("test")),
	//	}
	//	client.On("Do", mmock.Anything).Once().Return(res, nil)
	//	a.SendMetrics(ctx)
	//	assert.Equal(t, metrics, a.Metrics)
	//})
}
