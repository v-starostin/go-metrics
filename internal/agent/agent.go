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
	mu      *sync.Mutex
	l       *zerolog.Logger
	client  HTTPClient
	Metrics []model.AgentMetric
	address string
	counter *int64
	gw      *gzip.Writer
	//buf     *bytes.Buffer
}

func New(l *zerolog.Logger, c HTTPClient, a string) *Agent {
	counter := new(int64)
	*counter = 0
	return &Agent{
		mu:      &sync.Mutex{},
		l:       l,
		client:  c,
		address: a,
		counter: counter,
		gw:      gzip.NewWriter(io.Discard),
		//buf:     &bytes.Buffer{},
		Metrics: make([]model.AgentMetric, len(model.GaugeMetrics)+2),
	}
}

func (a *Agent) SendMetrics1(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(len(a.Metrics))

	for _, m := range a.Metrics {
		m := m
		go func() {
			defer wg.Done()
			b, err := json.Marshal(m)
			if err != nil {
				a.l.Error().Err(err).Msg("NewRequestWithContext method error")
				return
			}

			buf := &bytes.Buffer{}
			a.mu.Lock()
			//a.buf.Reset()
			a.gw.Reset(buf)
			//a.buf.Reset()
			n, err := a.gw.Write(b)
			if err != nil {
				return
			}
			a.gw.Close()
			a.mu.Unlock()

			a.l.Info().
				Int("len of b", len(b)).
				Int("written bytes", n).
				Int("len of buf", len(buf.Bytes())).
				Send()

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/update/", a.address), buf)
			if err != nil {
				a.l.Error().Err(err).Msg("NewRequestWithContext method error")
				return
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Encoding", "gzip")

			reqBody := req.Body
			reqBody.Close()

			r, err := gzip.NewReader(reqBody)
			if err != nil {
				a.l.Error().Err(err).Msg("gzip.NewReader error")
				return
			}

			reader := bytes.Buffer{}
			reader.ReadFrom(r)
			r.Close()

			a.l.Info().Msgf("request body is: %+v", reader.String())

			res, err := a.client.Do(req)
			if err != nil {
				a.l.Error().Err(err).Msg("Do method error")
				return
			}
			res.Body.Close()
		}()
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
	var gw *gzip.Writer
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))
	buf := &bytes.Buffer{}

	for _, m := range metrics {
		m := m

		go func() {
			defer wg.Done()
			b, err := json.Marshal(m)
			if err != nil {
				l.Error().Err(err).Msg("NewRequestWithContext method error")
				return
			}

			mu.Lock()
			gw = pool.Get().(*gzip.Writer)
			l.Info().Msgf("gw points to: %p", gw)
			gw.Reset(buf)
			buf.Reset()
			l.Info().Msgf("buffer points to: %p", buf)
			l.Info().Msgf("buffer's content: %s", (*buf).String())
			n, err := gw.Write(b)
			if err != nil {
				return
			}
			l.Info().Msgf("buffer's content: %s", (*buf).String())
			gw.Close()
			pool.Put(gw)

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
			mu.Unlock()
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

func (a *Agent) CollectMetrics1() {
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
