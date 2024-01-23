package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Agent struct {
	mu      sync.Mutex
	logger  *zerolog.Logger
	client  HTTPClient
	Metrics []model.AgentMetric
	address string
	counter *int64
	gw      *gzip.Writer
}

func New(logger *zerolog.Logger, client HTTPClient, address string) *Agent {
	counter := new(int64)
	*counter = 0
	return &Agent{
		logger:  logger,
		client:  client,
		address: address,
		counter: counter,
		gw:      gzip.NewWriter(io.Discard),
		Metrics: make([]model.AgentMetric, len(model.GaugeMetrics)+2),
	}
}

func (a *Agent) SendMetrics(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(len(a.Metrics))

	for _, m := range a.Metrics {
		go func(m model.AgentMetric) {
			defer wg.Done()
			b, err := json.Marshal(m)
			if err != nil {
				a.logger.Error().Err(err).Msg("Marshalling error")
				return
			}

			buf := &bytes.Buffer{}
			a.mu.Lock()
			a.gw.Reset(buf)
			n, err := a.gw.Write(b)
			if err != nil {
				a.logger.Error().Err(err).Msg("gw.Write error")
				return
			}
			a.gw.Close()
			a.mu.Unlock()

			a.logger.Info().
				Int("len of b", len(b)).
				Int("written bytes", n).
				Int("len of buf", len(buf.Bytes())).
				Send()

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/update/", a.address), buf)
			if err != nil {
				a.logger.Error().Err(err).Msg("http.NewRequestWithContext method error")
				return
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Encoding", "gzip")

			res, err := a.client.Do(req)
			if err != nil {
				a.logger.Error().Err(err).Msg("client.Do method error")
				return
			}
			res.Body.Close()
			a.logger.Info().Any("metric", m).Msg("Metric is sent")
		}(m)
	}
	wg.Wait()
}

func (a *Agent) CollectMetrics() {
	*a.counter++
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	msvalue := reflect.ValueOf(memStats)
	mstype := msvalue.Type()

	for index, metric := range model.GaugeMetrics {
		field, ok := mstype.FieldByName(metric)
		if !ok {
			continue
		}
		value := msvalue.FieldByName(metric).Interface()
		a.Metrics[index] = model.AgentMetric{MType: service.TypeGauge, ID: field.Name, Value: value}
	}
	a.Metrics[len(model.GaugeMetrics)] = model.AgentMetric{MType: service.TypeGauge, ID: "RandomValue", Value: rand.Float64()}
	a.Metrics[len(model.GaugeMetrics)+1] = model.AgentMetric{MType: service.TypeCounter, ID: "PollCount", Delta: *a.counter}
}
