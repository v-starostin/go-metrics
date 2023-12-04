package service

import (
	"log"
	"strconv"

	"github.com/pkg/errors"
)

const (
	TypeCounter = "counter"
	TypeGauge   = "gauge"
	CmdUpdate   = "update"
)

var (
	ErrParseMetric = errors.New("failed to parse metric: wrong type")
	ErrStoreData   = errors.New("failed to store data")
)

type Service struct {
	repo Repository
}

type Repository interface {
	StoreCounter(mtype, mname string, mvalue int64) bool
	StoreGauge(mtype, mname string, mvalue float64) bool
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Save(mtype, mname, mvalue string) error {
	if mtype == TypeCounter {
		log.Println("metrics type counter: parsing value string to int64")
		value, err := strconv.ParseInt(mvalue, 10, 0)
		if err != nil {
			return ErrParseMetric
		}

		log.Println("storing data to storage (counter)")
		if ok := s.repo.StoreCounter(mtype, mname, value); !ok {
			return ErrStoreData
		}

		return nil
	}

	if mtype == TypeGauge {
		log.Println("metrics type gauge: parsing value string to float64")
		value, err := strconv.ParseFloat(mvalue, 64)
		if err != nil {
			return ErrParseMetric
		}

		log.Println("storing data to storage (gauge)")
		if ok := s.repo.StoreGauge(mtype, mname, value); !ok {
			return ErrStoreData
		}

		return nil
	}

	return nil
}
