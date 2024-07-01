package service

import (
	"context"
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

const (
	maxRetries  = 3
	firstRetry  = 1 * time.Second
	secondRetry = 3 * time.Second
	thirdRetry  = 5 * time.Second
)

var ErrParseMetric = errors.New("failed to parse metric: wrong type")

// Repository defines methods for loading, storing, and managing metrics.
type Repository interface {
	Load(ctx context.Context, mtype, mname string) (*model.Metric, error)
	LoadAll(ctx context.Context) (model.Data, error)
	StoreMetric(ctx context.Context, m model.Metric) error
	StoreMetrics(ctx context.Context, m []model.Metric) error
	PingStorage(ctx context.Context) error
	RestoreFromFile() error
	WriteToFile() error
}

// Service represents a service for managing metrics.
type Service struct {
	logger *zerolog.Logger
	repo   Repository
}

// New creates a new Service with the provided logger and repository.
func New(l *zerolog.Logger, repo Repository) *Service {
	return &Service{
		logger: l,
		repo:   repo,
	}
}

// GetMetric retrieves a specific metric by its type and name.
func (s *Service) GetMetric(ctx context.Context, mtype, mname string) (*model.Metric, error) {
	var m *model.Metric
	var err error
	err = s.Retry(ctx, maxRetries, func(ctx context.Context) error {
		m, err = s.repo.Load(ctx, mtype, mname)
		if err != nil {
			return err
		}
		return nil
	}, firstRetry, secondRetry, thirdRetry)
	if err != nil {
		return nil, fmt.Errorf("failed to load metric %s: %w", mname, err)
	}

	return m, nil
}

// GetMetrics retrieves all metrics.
func (s *Service) GetMetrics(ctx context.Context) (model.Data, error) {
	var m model.Data
	var err error
	err = s.Retry(ctx, maxRetries, func(ctx context.Context) error {
		m, err = s.repo.LoadAll(ctx)
		if err != nil {
			return err
		}
		return nil
	}, firstRetry, secondRetry, thirdRetry)
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics: %w", err)
	}
	return m, nil
}

// SaveMetric saves a single metric.
func (s *Service) SaveMetric(ctx context.Context, m model.Metric) error {
	logger := s.logger.With().
		Str("type", m.MType).
		Str("name", m.ID).
		Logger()

	err := s.Retry(ctx, maxRetries, func(ctx context.Context) error {
		if err := s.repo.StoreMetric(ctx, m); err != nil {
			return err
		}
		return nil
	}, firstRetry, secondRetry, thirdRetry)
	if err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}
	logger.Info().Msg("Metric is stored")
	return nil
}

// SaveMetrics saves multiple metrics.
func (s *Service) SaveMetrics(ctx context.Context, m []model.Metric) error {
	err := s.Retry(ctx, maxRetries, func(ctx context.Context) error {
		if err := s.repo.StoreMetrics(ctx, m); err != nil {
			return err
		}
		return nil
	}, firstRetry, secondRetry, thirdRetry)
	if err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}
	s.logger.Info().Msg("Metric is stored")
	return nil
}

// PingStorage checks the connection to the storage.
func (s *Service) PingStorage(ctx context.Context) error {
	return s.repo.PingStorage(ctx)
}

// WriteToFile writes the current state to a file.
func (s *Service) WriteToFile() error {
	return s.repo.WriteToFile()
}

// RestoreFromFile restores the state from a file.
func (s *Service) RestoreFromFile() error {
	return s.repo.RestoreFromFile()
}

// Retry attempts to execute the given function up to a specified number of retries.
func (s *Service) Retry(ctx context.Context, maxRetries int, fn func(context.Context) error, intervals ...time.Duration) error {
	var err error
	err = fn(ctx)
	if err == nil {
		return nil
	}
	for i := 0; i < maxRetries; i++ {
		s.logger.Info().Msgf("Retrying... (Attempt %d)", i+1)
		time.Sleep(intervals[i])
		if err = fn(ctx); err == nil {
			return nil
		}
	}
	s.logger.Error().Msg("Retrying... Failed")
	return err
}
