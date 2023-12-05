package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	"github.com/v-starostin/go-metrics/internal/model"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func SendMetrics(ctx context.Context, client HTTPClient, metrics []model.Metric) error {
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))

	for _, m := range metrics {
		m := m

		go func() {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://0.0.0.0:8080/update/%s/%s/%v", m.Type, m.Name, m.Value), nil)
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/plain")
			res, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			res.Body.Close()
		}()
	}

	wg.Wait()

	return nil
}

func CollectMetrics() []model.Metric {
	var metrics []model.Metric
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	msvalue := reflect.ValueOf(memStats)
	mstype := msvalue.Type()

	for _, metric := range model.GaugeMetrics {
		field, ok := mstype.FieldByName(metric)
		if !ok {
			continue
		}
		value := msvalue.FieldByName(metric)
		metrics = append(metrics, model.Metric{Type: "gauge", Name: field.Name, Value: value})
	}
	fmt.Printf("metrics: %+v\n", metrics)

	return metrics
}
