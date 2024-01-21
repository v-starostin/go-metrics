package repository

import (
	"database/sql"
	"errors"

	"github.com/rs/zerolog"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

var errNotSupported = errors.New("not supported when DB is enabled")

type Storage struct {
	db     *sql.DB
	logger *zerolog.Logger
}

func NewStorage(logger *zerolog.Logger, db *sql.DB) *Storage {
	return &Storage{
		logger: logger,
		db:     db,
	}
}

func (s *Storage) Load(mtype, mname string) *model.Metric {
	var mID, mType string
	var mValue sql.NullFloat64
	var mDelta sql.NullInt64

	row := s.db.QueryRow("SELECT id, type, value, delta FROM metrics WHERE type = $1 AND id = $2", mtype, mname)
	if err := row.Scan(&mID, &mType, &mValue, &mDelta); err != nil {
		s.logger.Error().Err(err).Msg("Load method error")
		return nil
	}

	return &model.Metric{
		MType: mType,
		ID:    mID,
		Value: parseValue(mValue),
		Delta: parseDelta(mDelta),
	}
}

func (s *Storage) LoadAll() model.Data {
	rows, err := s.db.Query("SELECT id,type,value,delta FROM metrics")
	if err != nil {
		s.logger.Error().Err(err).Msg("LoadAll: select statement error")
		return nil
	}
	defer rows.Close()

	result := make(model.Data)
	for rows.Next() {
		var mID, mType string
		var mValue sql.NullFloat64
		var mDelta sql.NullInt64

		if err := rows.Scan(&mID, &mType, &mValue, &mDelta); err != nil {
			s.logger.Error().Err(err).Msg("LoadAll: scan rows error")
			return nil
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
		return nil
	}

	return result
}

func (s *Storage) StoreMetrics(metrics []model.Metric) bool {
	var stored bool

	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error().Err(err).Msg("StoreMetrics: begin transaction error")
		return false
	}

	for _, metric := range metrics {
		stored = store(tx, s.logger, metric)
		if !stored {
			s.logger.Error().Err(err).Msg("StoreMetrics: store data error")
			tx.Rollback()
			return false
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error().Err(err).Msg("StoreMetrics: commit transaction error")
		return false
	}

	return true
}

func store(tx *sql.Tx, logger *zerolog.Logger, m model.Metric) bool {
	logger.Info().Any("metric", m).Send()

	var mID, mType string
	var mDelta sql.NullInt64

	raw := tx.QueryRow("SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2", m.ID, m.MType)
	err := raw.Scan(&mID, &mType, &mDelta)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error().Err(err).Msg("store: scan row error")
		return false
	}

	if m.MType == service.TypeCounter {
		switch {
		case mID != "":
			_, err = tx.Exec(
				"UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
				mDelta.Int64+*m.Delta, m.ID, m.MType,
			)
		default:
			_, err = tx.Exec(
				"INSERT INTO metrics (id, type, delta) VALUES ($1,$2,$3)",
				m.ID, m.MType, *m.Delta,
			)
		}
		if err != nil {
			logger.Error().Err(err).Msg("store: error to store counter")
			return false
		}
	}

	if m.MType == service.TypeGauge {
		switch {
		case mID != "":
			_, err = tx.Exec(
				"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
				m.Value, m.ID, m.MType,
			)
		default:
			_, err = tx.Exec(
				"INSERT INTO metrics (id, type, value) VALUES ($1,$2,$3)",
				m.ID, m.MType, *m.Value,
			)
		}
		if err != nil {
			logger.Error().Err(err).Msg("store: error to store gauge")
			return false
		}
	}

	return true
}

func (s *Storage) Store(m model.Metric) bool {
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error().Err(err).Msg("Store: begin transaction error")
		return false
	}
	if stored := store(tx, s.logger, m); !stored {
		s.logger.Error().Err(err).Msg("Store: store data error")
		tx.Rollback()
		return false
	}
	if err := tx.Commit(); err != nil {
		s.logger.Error().Err(err).Msg("Store: commit transaction error")
		return false
	}
	return true
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
