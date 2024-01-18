package repository

import (
	"database/sql"
	"errors"

	"github.com/v-starostin/go-metrics/internal/model"
)

type DB struct {
	db *sql.DB
}

func NewDB() *DB {
	return &DB{}
}

func (db *DB) Load(mtype, mname string) *model.Metric {
	return nil
}

func (db *DB) LoadAll() model.Data {
	return nil
}

func (db *DB) Store(m model.Metric) bool {
	return false
}

func (db *DB) RestoreFromFile() error {
	return errors.New("err")
}

func (db *DB) WriteToFile() error {
	return errors.New("err")
}
