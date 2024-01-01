package service

import (
	"fmt"
	"strconv"

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
	StoreCounter(mtype, mname string, mvalue int64) bool
	StoreGauge(mtype, mname string, mvalue float64) bool
	Load(mtype, mname string) *model.Metric
	LoadAll() model.Data
}

func New(l *zerolog.Logger, repo Repository) *Service {
	return &Service{
		logger: l,
		repo:   repo,
	}
}

func (s *Service) GetMetric(mtype, mname string) (*model.Metric, error) {
	m := s.repo.Load(mtype, mname)
	if m == nil {
		return nil, fmt.Errorf("failed to load metric %s", mname)
	}

	return m, nil
}

func (s *Service) GetMetrics() (model.Data, error) {
	m := s.repo.LoadAll()
	if m == nil {
		return nil, errors.New("failed to load metrics")
	}

	return m, nil
}

func (s *Service) SaveMetric(mtype, mname, mvalue string) error {
	logger := s.logger.With().
		Str("type", mtype).
		Str("name", mname).
		Str("value", mvalue).
		Logger()

	switch mtype {
	case TypeCounter:
		value, err := strconv.ParseInt(mvalue, 10, 0)
		if err != nil {
			return ErrParseMetric
		}

		if ok := s.repo.StoreCounter(mtype, mname, value); !ok {
			return ErrStoreData
		}
		logger.Info().Msg("Metric is stored")
	case TypeGauge:
		value, err := strconv.ParseFloat(mvalue, 64)
		if err != nil {
			return ErrParseMetric
		}

		if ok := s.repo.StoreGauge(mtype, mname, value); !ok {
			return ErrStoreData
		}
		logger.Info().Msg("Metric is stored")
	}

	return nil
}
