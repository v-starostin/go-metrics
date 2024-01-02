package agent

import (
	"bytes"
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

func SendMetrics(ctx context.Context, l *zerolog.Logger, client HTTPClient, metrics []model.Metric, address string) error {
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
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/update/", address), bytes.NewReader(b))
			if err != nil {
				l.Error().Err(err).Msg("NewRequestWithContext method error")
				return
			}
			req.Header.Add("Content-Type", "application/json")
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

func CollectMetrics(metrics []model.Metric, counter *int64) {
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
		value := msvalue.FieldByName(metric).Float()
		metrics[index] = model.Metric{MType: service.TypeGauge, ID: field.Name, Value: &value}
	}
	r := rand.Float64()
	metrics[len(model.GaugeMetrics)] = model.Metric{MType: service.TypeGauge, ID: "RandomValue", Value: &r}
	metrics[len(model.GaugeMetrics)+1] = model.Metric{MType: service.TypeCounter, ID: "PollCount", Delta: counter}
}
