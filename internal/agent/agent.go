package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func SendMetrics(ctx context.Context, client HTTPClient, metrics []model.Metric, address string) error {
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))

	for _, m := range metrics {
		m := m

		go func() {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/update/%s/%s/%v", address, m.Type, m.Name, m.Value), nil)
			if err != nil {
				log.Println(err)
				return
			}
			req.Header.Add("Content-Type", "text/plain")
			res, err := client.Do(req)
			if err != nil {
				log.Println(err)
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
		value := msvalue.FieldByName(metric)
		metrics[index] = model.Metric{Type: service.TypeGauge, Name: field.Name, Value: value}
	}
	metrics[len(model.GaugeMetrics)] = model.Metric{Type: service.TypeGauge, Name: "RandomValue", Value: rand.Float64()}
	metrics[len(model.GaugeMetrics)+1] = model.Metric{Type: service.TypeCounter, Name: "PollCount", Value: *counter}
}
