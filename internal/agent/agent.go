package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
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
				l.Error().Err(err).Msg("NewRequestWithContext method error")
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

func CollectMetrics(metrics []model.AgentMetric, counter *int64) {
	*counter++
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	msvalue := reflect.ValueOf(memStats)
	mstype := msvalue.Type()

	for _, metric := range model.GaugeMetrics {
		field, ok := mstype.FieldByName(metric)
		if !ok {
			continue
		}
		value := msvalue.FieldByName(metric).Interface()
		metrics = append(metrics, model.AgentMetric{MType: service.TypeGauge, ID: field.Name, Value: value})
	}
	metrics = append(metrics, model.AgentMetric{MType: service.TypeGauge, ID: "RandomValue", Value: rand.Float64()})
	metrics = append(metrics, model.AgentMetric{MType: service.TypeCounter, ID: "PollCount", Delta: *counter})
}
