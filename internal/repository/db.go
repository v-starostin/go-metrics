package repository

import (
	"database/sql"
	"errors"
	"log"

	"github.com/v-starostin/go-metrics/internal/model"
	"github.com/v-starostin/go-metrics/internal/service"
)

var errNotSupported = errors.New("not supported when DB is enabled")

type DB struct {
	*sql.DB
}

func NewDB(db *sql.DB) *DB {
	return &DB{db}
}

func (db *DB) Load(mtype, mname string) *model.Metric {
	var mID, mType string
	var mValue sql.NullFloat64
	var mDelta sql.NullInt64

	row := db.QueryRow("SELECT id, type, value, delta FROM metrics WHERE type = $1 AND id = $2", mtype, mname)
	if err := row.Scan(&mID, &mType, &mValue, &mDelta); err != nil {
		return nil
	}

	return &model.Metric{
		MType: mType,
		ID:    mID,
		Value: parseValue(mValue),
		Delta: parseDelta(mDelta),
	}
}

func (db *DB) LoadAll() model.Data {
	rows, err := db.Query("SELECT id,type,value,delta FROM metrics")
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make(model.Data)
	for rows.Next() {
		var mID, mType string
		var mValue sql.NullFloat64
		var mDelta sql.NullInt64

		if err := rows.Scan(&mID, &mType, &mValue, &mDelta); err != nil {
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
		return nil
	}

	return result
}

func (db *DB) StoreMetrics(metrics []model.Metric) bool {
	var stored bool
	tx, err := db.Begin()
	if err != nil {
		log.Println("StoreMetrics err :", err.Error())
		return false
	}
	for _, metric := range metrics {
		stored = store(tx, metric)
		if !stored {
			log.Println("Failed to store data")
			tx.Rollback()
			return false
		}
	}
	if err := tx.Commit(); err != nil {
		log.Println("Error to commit transaction", err.Error())
	}
	return true
}

func store(tx *sql.Tx, m model.Metric) bool {
	log.Printf("metric: %+v\n", m)
	var mID, mType string
	var mDelta sql.NullInt64

	raw := tx.QueryRow("SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2", m.ID, m.MType)
	if err := raw.Scan(&mID, &mType, &mDelta); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("[1] err:", err.Error())
		return false
	}

	if mID != "" && m.MType == service.TypeCounter {
		result, err := tx.Exec(
			"UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
			mDelta.Int64+*m.Delta, m.ID, m.MType,
		)
		if err != nil {
			log.Println("[2] err:", err.Error())
			return false
		}
		affected, err := result.RowsAffected()
		if err != nil {
			log.Println("[3] err:", err.Error())
			return false
		}
		if affected != 1 {
			return false
		}
		return true
	}

	if mID != "" && m.MType == service.TypeGauge {
		result, err := tx.Exec(
			"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
			m.Value, m.ID, m.MType,
		)
		if err != nil {
			log.Println("[4] err:", err.Error())
			return false
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return false
		}
		if affected != 1 {
			return false
		}
		return true
	}

	var result sql.Result
	var err error
	if m.MType == service.TypeGauge {
		result, err = tx.Exec(
			"INSERT INTO metrics (id, type, value) VALUES ($1,$2,$3)",
			m.ID, m.MType, *m.Value,
		)
	} else {
		result, err = tx.Exec(
			"INSERT INTO metrics (id, type, delta) VALUES ($1,$2,$3)",
			m.ID, m.MType, *m.Delta,
		)
	}
	if err != nil {
		log.Println("[6] err:", err.Error())
		return false
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false
	}
	if affected != 1 {
		return false
	}

	return true
}

func (db *DB) Store(m model.Metric) bool {
	tx, err := db.Begin()
	if err != nil {
		log.Println("store err:", err.Error())
		return false
	}
	if stored := store(tx, m); !stored {
		log.Println("store err")
		tx.Rollback()
		return false
	}
	if err := tx.Commit(); err != nil {
		log.Println("Error to commit transaction", err.Error())
		return false
	}
	return true
}

func (db *DB) Store1(m model.Metric) bool {
	var mID, mType string
	var mDelta sql.NullInt64

	raw := db.QueryRow("SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2", m.ID, m.MType)
	if err := raw.Scan(&mID, &mType, &mDelta); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false
	}

	if mID != "" && m.MType == service.TypeCounter {
		result, err := db.Exec(
			"UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
			mDelta.Int64+*m.Delta, m.ID, m.MType,
		)
		if err != nil {
			return false
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return false
		}
		if affected != 1 {
			return false
		}
		return true
	}

	if mID != "" && m.MType == service.TypeGauge {
		result, err := db.Exec(
			"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
			m.Value, m.ID, m.MType,
		)
		if err != nil {
			return false
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return false
		}
		if affected != 1 {
			return false
		}
		return true
	}

	var result sql.Result
	var err error
	if m.MType == service.TypeGauge {
		result, err = db.Exec(
			"INSERT INTO metrics (id, type, value) VALUES ($1,$2,$3)",
			m.ID, m.MType, *m.Value,
		)
	} else {
		result, err = db.Exec(
			"INSERT INTO metrics (id, type, delta) VALUES ($1,$2,$3)",
			m.ID, m.MType, *m.Delta,
		)
	}
	if err != nil {
		return false
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false
	}
	if affected != 1 {
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

func (db *DB) RestoreFromFile() error {
	return errNotSupported
}

func (db *DB) WriteToFile() error {
	return errNotSupported
}
