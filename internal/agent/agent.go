package agent

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/v-starostin/go-metrics/internal/model"
)

func SendMetrics(ctx context.Context, metrics []model.Metric) error {
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))
	client := http.Client{Timeout: 3 * time.Minute}

	for _, m := range metrics {
		m := m

		go func() error {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://0.0.0.0:8080/update/%s/%s/%v", m.Type, m.Name, m.Value), nil)
			if err != nil {
				return err
			}
			req.Header.Add("Content-Type", "text/plain")
			res, err := client.Do(req)
			if err != nil {
				return err
			}
			res.Body.Close()
			return nil
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
