package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

var errNotSupported = errors.New("not supported when DB is enabled")

// Storage represents a storage
type Storage struct {
	db     *sql.DB
	logger *zerolog.Logger
}

// NewStorage creates a new Storage.
func NewStorage(logger *zerolog.Logger, db *sql.DB) *Storage {
	return &Storage{
		logger: logger,
		db:     db,
	}
}

// Load retrieves a specific metric by its type and name from the database.
func (s *Storage) Load(ctx context.Context, mtype, mname string) (*model.Metric, error) {
	var mID, mType string
	var mValue sql.NullFloat64
	var mDelta sql.NullInt64

	row := s.db.QueryRowContext(ctx, "SELECT id, type, value, delta FROM metrics WHERE type = $1 AND id = $2", mtype, mname)
	if err := row.Scan(&mID, &mType, &mValue, &mDelta); err != nil {
		s.logger.Error().Err(err).Msg("Load method error")
		return nil, err
	}

	return &model.Metric{
		MType: mType,
		ID:    mID,
		Value: parseValue(mValue),
		Delta: parseDelta(mDelta),
	}, nil
}

// LoadAll retrieves all metrics from the database.
func (s *Storage) LoadAll(ctx context.Context) (model.Data, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id,type,value,delta FROM metrics")
	if err != nil {
		s.logger.Error().Err(err).Msg("LoadAll: select statement error")
		return nil, err
	}
	defer rows.Close()

	result := make(model.Data)
	for rows.Next() {
		var mID, mType string
		var mValue sql.NullFloat64
		var mDelta sql.NullInt64

		if err := rows.Scan(&mID, &mType, &mValue, &mDelta); err != nil {
			s.logger.Error().Err(err).Msg("LoadAll: scan rows error")
			return nil, err
		}
		_, ok := result[mType]
		if !ok {
			result[mType] = map[string]model.Metric{
				mID: {
					ID:    mID,
					MType: mType,
					Delta: parseDelta(mDelta),
					Value: parseValue(mValue),
				},
			}
			continue
		}
		result[mType][mID] = model.Metric{
			ID:    mID,
			MType: mType,
			Delta: parseDelta(mDelta),
			Value: parseValue(mValue),
		}
	}
	if err := rows.Err(); err != nil {
		s.logger.Error().Err(err).Msg("LoadAll method error")
		return nil, err
	}

	return result, nil
}

// StoreMetrics saves multiple metrics to the database.
func (s *Storage) StoreMetrics(ctx context.Context, metrics []model.Metric) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("StoreMetrics: begin transaction error")
		return err
	}

	for _, metric := range metrics {
		err = store(ctx, tx, s.logger, metric)
		if err != nil {
			s.logger.Error().Err(err).Msg("StoreMetrics: store data error")
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error().Err(err).Msg("StoreMetrics: commit transaction error")
		return err
	}

	return nil
}

func store(ctx context.Context, tx *sql.Tx, logger *zerolog.Logger, m model.Metric) error {
	logger.Info().Any("metric", m).Send()

	var mID, mType string
	var mDelta sql.NullInt64

	raw := tx.QueryRowContext(ctx, "SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2", m.ID, m.MType)
	err := raw.Scan(&mID, &mType, &mDelta)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error().Err(err).Msg("store: scan row error")
		return err
	}

	if m.MType == service.TypeCounter {
		switch {
		case mID != "":
			_, err = tx.ExecContext(
				ctx,
				"UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
				mDelta.Int64+*m.Delta, m.ID, m.MType,
			)
		default:
			_, err = tx.ExecContext(
				ctx,
				"INSERT INTO metrics (id, type, delta) VALUES ($1,$2,$3)",
				m.ID, m.MType, *m.Delta,
			)
		}
		if err != nil {
			logger.Error().Err(err).Msg("store: error to store counter")
			return err
		}
	}

	if m.MType == service.TypeGauge {
		switch {
		case mID != "":
			_, err = tx.ExecContext(
				ctx,
				"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
				m.Value, m.ID, m.MType,
			)
		default:
			_, err = tx.ExecContext(
				ctx,
				"INSERT INTO metrics (id, type, value) VALUES ($1,$2,$3)",
				m.ID, m.MType, *m.Value,
			)
		}
		if err != nil {
			logger.Error().Err(err).Msg("store: error to store gauge")
			return err
		}
	}

	return nil
}

// StoreMetric saves a single metric to the database.
func (s *Storage) StoreMetric(ctx context.Context, m model.Metric) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("Store: begin transaction error")
		return err
	}
	if err := store(ctx, tx, s.logger, m); err != nil {
		s.logger.Error().Err(err).Msg("Store: store data error")
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		s.logger.Error().Err(err).Msg("Store: commit transaction error")
		return err
	}
	return nil
}

// PingStorage checks the connection to the storage.
func (s *Storage) PingStorage(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func parseDelta(v sql.NullInt64) *int64 {
	if v.Valid {
		return &v.Int64
	}
	return nil
}

func parseValue(v sql.NullFloat64) *float64 {
	if v.Valid {
		return &v.Float64
	}
	return nil
}

func (s *Storage) RestoreFromFile() error {
	return errNotSupported
}

func (s *Storage) WriteToFile() error {
	return errNotSupported
}
