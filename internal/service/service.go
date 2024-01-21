package service

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
)

const (
	TypeCounter = "counter"
	TypeGauge   = "gauge"
)

var (
	ErrParseMetric = errors.New("failed to parse metric: wrong type")
	ErrStoreData   = errors.New("failed to store data")
)

type Service struct {
	logger *zerolog.Logger
	repo   Repository
}

type Repository interface {
	Load(mtype, mname string) *model.Metric
	LoadAll() model.Data
	Store(m model.Metric) bool
	StoreMetrics(m []model.Metric) bool
	RestoreFromFile() error
	WriteToFile() error
}

func New(l *zerolog.Logger, repo Repository) *Service {
	return &Service{
		logger: l,
		repo:   repo,
	}
}

func (s *Service) GetMetric(mtype, mname string) (*model.Metric, error) {
	var m *model.Metric
	err := s.Retry(3, func() bool {
		m = s.repo.Load(mtype, mname)
		return m != nil
	}, 1*time.Second, 3*time.Second, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to load metric %s", mname)
	}

	return m, nil
}

func (s *Service) GetMetrics() (model.Data, error) {
	var m model.Data
	err := s.Retry(3, func() bool {
		m = s.repo.LoadAll()
		return m != nil
	}, 1*time.Second, 3*time.Second, 5*time.Second)
	if err != nil {
		return nil, errors.New("failed to load metrics")
	}

	return m, nil
}

func (s *Service) SaveMetric(m model.Metric) error {
	logger := s.logger.With().
		Str("type", m.MType).
		Str("name", m.ID).
		Logger()

	if ok := s.repo.Store(m); !ok {
		return ErrStoreData
	}
	logger.Info().Msg("Metric is stored")

	return nil
}

func (s *Service) SaveMetrics(m []model.Metric) error {
	//logger := s.logger.With().
	//	Str("type", m.MType).
	//	Str("name", m.ID).
	//	Logger()

	if ok := s.repo.StoreMetrics(m); !ok {
		return ErrStoreData
	}
	s.logger.Info().Msg("Metric is stored")

	return nil
}

func (s *Service) Retry(maxRetries int, fn func() bool, intervals ...time.Duration) error {
	var ok bool
	ok = fn()
	if ok {
		return nil
	}
	for i := 0; i < maxRetries; i++ {
		s.logger.Info().Msgf("Retrying... (Attempt %d)", i+1)
		time.Sleep(intervals[i])
		if ok = fn(); ok {
			return nil
		}
	}
	s.logger.Error().Msg("Retrying... Failed")
	return errors.New("err")
}
