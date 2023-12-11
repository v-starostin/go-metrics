package service

import (
	"fmt"
	"log"
	"strconv"

	"github.com/pkg/errors"

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
	repo Repository
}

type Repository interface {
	StoreCounter(mtype, mname string, mvalue int64) bool
	StoreGauge(mtype, mname string, mvalue float64) bool
	Load(mtype, mname string) *model.Metric
	LoadAll() model.Data
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Metric(mtype, mname string) (*model.Metric, error) {
	m := s.repo.Load(mtype, mname)
	if m == nil {
		return nil, fmt.Errorf("failed to load metric %s", mname)
	}

	return m, nil
}

func (s *Service) Metrics() (model.Data, error) {
	m := s.repo.LoadAll()
	if m == nil {
		return nil, errors.New("failed to load metrics")
	}

	return m, nil
}

func (s *Service) Save(mtype, mname, mvalue string) error {
	switch mtype {
	case TypeCounter:
		log.Println("metrics type counter: parsing value string to int64")
		value, err := strconv.ParseInt(mvalue, 10, 0)
		if err != nil {
			return ErrParseMetric
		}

		log.Println("storing data to storage (counter)")
		if ok := s.repo.StoreCounter(mtype, mname, value); !ok {
			return ErrStoreData
		}
	case TypeGauge:
		log.Println("metrics type gauge: parsing value string to float64")
		value, err := strconv.ParseFloat(mvalue, 64)
		if err != nil {
			return ErrParseMetric
		}

		log.Println("storing data to storage (gauge)")
		if ok := s.repo.StoreGauge(mtype, mname, value); !ok {
			return ErrStoreData
		}
	}

	return nil

	// if mtype == TypeCounter {
	// 	log.Println("metrics type counter: parsing value string to int64")
	// 	value, err := strconv.ParseInt(mvalue, 10, 0)
	// 	if err != nil {
	// 		return ErrParseMetric
	// 	}

	// 	log.Println("storing data to storage (counter)")
	// 	if ok := s.repo.StoreCounter(mtype, mname, value); !ok {
	// 		return ErrStoreData
	// 	}
	// 	return nil
	// }

	// if mtype == TypeGauge {
	// 	log.Println("metrics type gauge: parsing value string to float64")
	// 	value, err := strconv.ParseFloat(mvalue, 64)
	// 	if err != nil {
	// 		return ErrParseMetric
	// 	}

	// 	log.Println("storing data to storage (gauge)")
	// 	if ok := s.repo.StoreGauge(mtype, mname, value); !ok {
	// 		return ErrStoreData
	// 	}
	// 	return nil
	// }
}
