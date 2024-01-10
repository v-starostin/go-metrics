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
	logger  *zerolog.Logger
	client  HTTPClient
	address string
	Counter *int64
	Metrics []model.AgentMetric
	pool    *sync.Pool
}

func New(logger *zerolog.Logger, client HTTPClient, metrics []model.AgentMetric, address string) *Agent {
	counter := int64(0)
	return &Agent{
		logger:  logger,
		client:  client,
		address: address,
		Counter: &counter,
		Metrics: metrics,
		pool: &sync.Pool{
			New: func() any { return gzip.NewWriter(io.Discard) },
		},
	}
}

func (a *Agent) sendMetric(ctx context.Context, wg *sync.WaitGroup, m model.AgentMetric) {
	defer wg.Done()

	b, err := json.Marshal(m)
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to marshal")
		return
	}

	buf := &bytes.Buffer{}
	gw := a.pool.Get().(*gzip.Writer)
	defer a.pool.Put(gw)
	gw.Reset(buf)
	a.logger.Info().Msgf("buffer points to: %p", buf)
	a.logger.Info().Msgf("buffer's content: %v", buf.String())
	n, err := gw.Write(b)
	if err != nil {
		return
	}
	a.logger.Info().Msgf("buffer's content: %v", buf.String())
	gw.Close()

	a.logger.Info().
		Int("len of b", len(b)).
		Int("written bytes", n).
		Int("len of buf", len(buf.Bytes())).
		Send()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/update/", a.address), buf)
	if err != nil {
		a.logger.Error().Err(err).Msg("NewRequestWithContext method error")
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")
	res, err := a.client.Do(req)
	if err != nil {
		a.logger.Error().Err(err).Msg("Do method error")
		return
	}
	res.Body.Close()
}

func (a *Agent) SendMetrics1(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(len(a.Metrics))

	for _, m := range a.Metrics {
		//m := m
		go a.sendMetric(ctx, &wg, m)
	}
	wg.Wait()
}

func SendMetrics(
	ctx context.Context,
	l *zerolog.Logger,
	client HTTPClient,
	metrics []model.AgentMetric,
	address string,
	pool *sync.Pool,
) error {
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))

	for _, m := range metrics {
		m := m

		go func() {
			defer wg.Done()

			b, err := json.Marshal(m)
			if err != nil {
				l.Error().Err(err).Msg("Failed to marshal")
				return
			}

			buf := &bytes.Buffer{}
			gw := pool.Get().(*gzip.Writer)
			defer pool.Put(gw)
			gw.Reset(buf)
			l.Info().Msgf("buffer points to: %p", buf)
			l.Info().Msgf("buffer's content: %v", buf.String())
			n, err := gw.Write(b)
			if err != nil {
				return
			}
			l.Info().Msgf("buffer's content: %v", buf.String())
			gw.Close()

			l.Info().
				Int("len of b", len(b)).
				Int("written bytes", n).
				Int("len of buf", len(buf.Bytes())).
				Send()

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/update/", address), buf)
			if err != nil {
				l.Error().Err(err).Msg("NewRequestWithContext method error")
				return
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Encoding", "gzip")
			res, err := client.Do(req)
			if err != nil {
				l.Error().Err(err).Msg("Do method error")
				return
			}
			res.Body.Close()
		}()
	}
	wg.Wait()
	return nil
}

func (a *Agent) CollectMetrics1() {
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
	a.Metrics[len(model.GaugeMetrics)+1] = model.AgentMetric{MType: service.TypeCounter, ID: "PollCount", Delta: *a.Counter}
}

func CollectMetrics(metrics []model.AgentMetric, counter *int64) {
	*counter++
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
		metrics[index] = model.AgentMetric{MType: service.TypeGauge, ID: field.Name, Value: value}
	}
	metrics[len(model.GaugeMetrics)] = model.AgentMetric{MType: service.TypeGauge, ID: "RandomValue", Value: rand.Float64()}
	metrics[len(model.GaugeMetrics)+1] = model.AgentMetric{MType: service.TypeCounter, ID: "PollCount", Delta: *counter}
}
