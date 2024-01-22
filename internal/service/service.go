package service

import (
	"fmt"

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
	Load(mtype, mname string) (*model.Metric, error)
	LoadAll() (model.Data, error)
	Store(m model.Metric) error
	StoreMetrics(m []model.Metric) error
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
	m, err := s.repo.Load(mtype, mname)
	if err != nil {
		return nil, fmt.Errorf("failed to load metric %s: %w", mname, err)
	}

	return m, nil
}

func (s *Service) GetMetrics() (model.Data, error) {
	m, err := s.repo.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics: %w", err)
	}

	return m, nil
}

func (s *Service) SaveMetric(m model.Metric) error {
	logger := s.logger.With().
		Str("type", m.MType).
		Str("name", m.ID).
		Logger()

	if err := s.repo.Store(m); err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}
	logger.Info().Msg("Metric is stored")

	return nil
}

func (s *Service) SaveMetrics(m []model.Metric) error {
	if err := s.repo.StoreMetrics(m); err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}
	s.logger.Info().Msg("Metric is stored")

	return nil
}

//func (s *Service) Retry(maxRetries int, fn func() bool, intervals ...time.Duration) error {
//	var ok bool
//	ok = fn()
//	if ok {
//		return nil
//	}
//	for i := 0; i < maxRetries; i++ {
//		s.logger.Info().Msgf("Retrying... (Attempt %d)", i+1)
//		time.Sleep(intervals[i])
//		if ok = fn(); ok {
//			return nil
//		}
//	}
//	s.logger.Error().Msg("Retrying... Failed")
//	return errors.New("err")
//}
