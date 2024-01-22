package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

type MemStorage struct {
	mu              sync.RWMutex
	logger          *zerolog.Logger
	data            model.Data
	interval        int
	storageFileName string
}

func NewMemStorage(logger *zerolog.Logger, interval int, file string) *MemStorage {
	return &MemStorage{
		logger:          logger,
		interval:        interval,
		storageFileName: file,
		data:            make(model.Data),
	}
}

func (s *MemStorage) RestoreFromFile() error {
	_, err := os.Stat(s.storageFileName)
	if errors.Is(err, os.ErrNotExist) {
		return os.ErrNotExist
	}
	if err != nil {
		return err
	}
	b, err := os.ReadFile(s.storageFileName)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &s.data); err != nil {
		return err
	}
	s.logger.Info().Msgf("RestoreFromFile: %+v", s.data)
	return nil
}

func (s *MemStorage) WriteToFile() error {
	file, err := os.OpenFile(s.storageFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	s.logger.Info().Msg("File successfully opened")

	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	s.logger.Info().Msgf("Data successfully marshalled: %v", string(b))

	n, err := file.Write(b)
	if err != nil {
		return err
	}
	s.logger.Info().Msgf("%d bytes were written to the file", n)
	return nil
}

func (s *MemStorage) LoadAll() (model.Data, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data, nil
}

func (s *MemStorage) Load(mtype, mname string) (*model.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics, ok := s.data[mtype]
	if !ok {
		return nil, fmt.Errorf("metric type: %s does not exist", mtype)
	}

	mvalue, ok := metrics[mname]
	if !ok {
		return nil, fmt.Errorf("metric name: %s does not exist", mname)
	}

	return &mvalue, nil
}

func (s *MemStorage) Store(m model.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.interval == 0 {
		defer func() {
			if err := s.WriteToFile(); err != nil {
				s.logger.Error().Err(err).Msg("Failed to write storage content to file")
			}
		}()
	}

	metrics, ok := s.data[m.MType]
	if !ok {
		s.data[m.MType] = map[string]model.Metric{
			m.ID: {ID: m.ID, MType: m.MType, Value: m.Value, Delta: m.Delta},
		}
		return nil
	}

	switch m.MType {
	case service.TypeGauge:
		metrics[m.ID] = model.Metric{ID: m.ID, MType: m.MType, Value: m.Value}
	case service.TypeCounter:
		metric, ok := metrics[m.ID]
		if !ok {
			metrics[m.ID] = model.Metric{ID: m.ID, MType: m.MType, Delta: m.Delta}
			return nil
		}
		*metric.Delta += *m.Delta
		metrics[m.ID] = model.Metric{ID: m.ID, MType: m.MType, Delta: metric.Delta}
	}
	s.logger.Info().Interface("Storage content", s.data).Send()

	return nil
}

func (s *MemStorage) StoreMetrics(metrics []model.Metric) error {
	var err error
	for _, metric := range metrics {
		err = s.Store(metric)
		if err != nil {
			return err
		}
	}
	return nil
}
