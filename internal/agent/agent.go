package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/v-starostin/go-metrics/internal/crypto"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

// HTTPClient defines a method for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Agent represents an agent that collects and sends metrics.
type Agent struct {
	mu        sync.Mutex
	logger    *zerolog.Logger
	client    HTTPClient
	Metrics   []model.AgentMetric
	address   string
	key       string
	counter   *int64
	gw        *gzip.Writer
	publicKey *rsa.PublicKey
}

// New creates a new Agent with the provided logger, HTTP client, address, and key.
func New(logger *zerolog.Logger, client HTTPClient, address, key string, publicKey *rsa.PublicKey) *Agent {
	counter := new(int64)
	*counter = 0
	return &Agent{
		logger:    logger,
		client:    client,
		address:   address,
		key:       key,
		counter:   counter,
		gw:        gzip.NewWriter(io.Discard),
		Metrics:   make([]model.AgentMetric, len(model.GaugeMetrics)+5),
		publicKey: publicKey,
	}
}

// SendMetrics sends the collected metrics to the configured address.
// It reads metrics from the provided channel and sends them in a compressed JSON format.
// If an error occurs during the process, it is logged and returned.
func (a *Agent) SendMetrics(ctx context.Context, metrics <-chan []model.AgentMetric) error {
	for {
		var m model.AgentMetrics
		var ok bool
		m, ok = <-metrics
		if !ok {
			return nil
		} else {
			//b, err := json.Marshal(m)
			b, err := m.MarshalJSON()
			if err != nil {
				a.logger.Error().Err(err).Msg("Marshalling error")
				return err
			}
			a.logger.Info().Any("json", string(b)).Msg("Marshalled")

			buf := &bytes.Buffer{}
			a.gw.Reset(buf)
			n, err := a.gw.Write(b)
			if err != nil {
				a.logger.Error().Err(err).Msg("gw.Write error")
				return err
			}
			a.gw.Close()

			a.logger.Info().
				Int("len of b", len(b)).
				Int("written bytes", n).
				Int("len of buf", len(buf.Bytes())).
				Send()

			var req *http.Request
			if a.publicKey != nil {
				encrypted, err := crypto.RSAEncrypt(a.publicKey, buf.Bytes())
				if err != nil {
					a.logger.Error().Err(err).Msg("Error to encrypt data")
					return err
				}
				req, err = http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/updates/", a.address), bytes.NewReader(encrypted))
				if err != nil {
					a.logger.Error().Err(err).Msg("http.NewRequestWithContext method error")
					return err
				}
			} else {
				req, err = http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/updates/", a.address), buf)
				if err != nil {
					a.logger.Error().Err(err).Msg("http.NewRequestWithContext method error")
					return err
				}
			}

			if a.key != "" {
				buf2 := *buf
				h := hmac.New(sha256.New, []byte(a.key))
				if _, err = h.Write(buf2.Bytes()); err != nil {
					return err
				}
				d := h.Sum(nil)
				a.logger.Info().Msgf("hash: %x", d)
				req.Header.Add("HashSHA256", hex.EncodeToString(d))
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Encoding", "gzip")

			res, err := a.client.Do(req)
			if err != nil {
				a.logger.Error().Err(err).Msg("client.Do method error")
				return err
			}
			res.Body.Close()
			a.logger.Info().Any("metric", m).Msg("Metrics are sent")
		}
	}
}

// CollectRuntimeMetrics collects runtime metrics at a specified interval and stores them in the Agent's Metrics slice.
func (a *Agent) CollectRuntimeMetrics(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)

	for {
		select {
		case <-t.C:
			atomic.AddInt64(a.counter, 1)
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			msvalue := reflect.ValueOf(memStats)
			mstype := msvalue.Type()

			for index, metric := range model.GaugeMetrics {
				field, ok := mstype.FieldByName(metric)
				if !ok {
					return
				}
				value := msvalue.FieldByName(metric).Interface()
				a.Metrics[index] = model.AgentMetric{MType: service.TypeGauge, ID: field.Name, Value: value}
			}

			a.Metrics[len(model.GaugeMetrics)] = model.AgentMetric{MType: service.TypeGauge, ID: "RandomValue", Value: rand.Float64()}
			a.Metrics[len(model.GaugeMetrics)+1] = model.AgentMetric{MType: service.TypeCounter, ID: "PollCount", Delta: *a.counter}
			a.logger.Info().Any("metric (collect)", a.Metrics).Msg("metric (collect)")
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

// CollectGopsutilMetrics collects Gopsutil metrics at a specified interval and stores them in the Agent's Metrics slice.
func (a *Agent) CollectGopsutilMetrics(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)

	for {
		select {
		case <-t.C:
			v, err := mem.VirtualMemory()
			if err != nil {
				return
			}
			a.Metrics[len(model.GaugeMetrics)+2] = model.AgentMetric{MType: service.TypeGauge, ID: "TotalMemory", Value: int64(v.Total)}
			a.Metrics[len(model.GaugeMetrics)+3] = model.AgentMetric{MType: service.TypeGauge, ID: "FreeMemory", Value: int64(v.Free)}
			a.Metrics[len(model.GaugeMetrics)+4] = model.AgentMetric{MType: service.TypeGauge, ID: "CPUutilization1", Value: v.UsedPercent}
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

// PrepareMetrics sends metrics to a channel.
func (a *Agent) PrepareMetrics(ctx context.Context, interval time.Duration) <-chan []model.AgentMetric {
	ch := make(chan []model.AgentMetric)
	wg := &sync.WaitGroup{}
	t := time.NewTicker(interval)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-t.C:
				ch <- a.Metrics
			case <-ctx.Done():
				t.Stop()
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

// Retry attempts to execute the given function up to a specified number of retries, with specified intervals between each retry.
func (a *Agent) Retry(ctx context.Context, maxRetries int, fn func(ctx context.Context) error, intervals ...time.Duration) error {
	var err error
	err = fn(ctx)
	if err == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for i := 0; i < maxRetries; i++ {
		a.logger.Info().Msgf("Retrying... (Attempt %d)", i+1)

		t := time.NewTimer(intervals[i])
		select {
		case <-t.C:
			if err = fn(ctx); err == nil {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}

	}
	a.logger.Error().Err(err).Msg("Retrying... Failed")
	return err
}
