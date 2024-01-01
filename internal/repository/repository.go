package repository

import (
	"sync"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
)

type MemStorage struct {
	mu     sync.RWMutex
	logger *zerolog.Logger
	data   model.Data
}

func New(l *zerolog.Logger) *MemStorage {
	return &MemStorage{
		logger: l,
		data:   make(model.Data),
	}
}

func (s *MemStorage) LoadAll() model.Data {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

func (s *MemStorage) Load(mtype, mname string) *model.Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics, ok := s.data[mtype]
	if !ok {
		s.logger.Info().Msgf("Metric type %s doesn't exist", mtype)
		return nil
	}

	mvalue, ok := metrics[mname]
	if !ok {
		s.logger.Info().Msgf("Metric %s doesn't exist", mvalue)
		return nil
	}

	return &mvalue
}

func (s *MemStorage) StoreCounter(mtype, mname string, mvalue int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	counter, ok := s.data[mtype]
	if !ok {
		s.data[mtype] = map[string]model.Metric{
			mname: {Name: mname, Type: mtype, Value: mvalue},
		}
		return true
	}
	val, ok := counter[mname]
	if !ok {
		counter[mname] = model.Metric{Name: mname, Type: mtype, Value: mvalue}
		return true
	}
	v := val.Value.(int64)
	v += mvalue
	counter[mname] = model.Metric{Name: mname, Type: mtype, Value: v}
	s.logger.Info().Interface("Storage content", s.data).Send()

	return true
}

func (s *MemStorage) StoreGauge(mtype, mname string, mvalue float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	gauge, ok := s.data[mtype]
	if !ok {
		s.data[mtype] = map[string]model.Metric{
			mname: {Name: mname, Type: mtype, Value: mvalue},
		}
		return true
	}

	gauge[mname] = model.Metric{Name: mname, Type: mtype, Value: mvalue}
	s.logger.Info().Interface("Storage content", s.data).Send()

	return true
}
