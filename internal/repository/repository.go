package repository

import (
	"log"

	"github.com/v-starostin/go-metrics/internal/model"
)

// type Data map[string]map[string]any

type MemStorage struct {
	data model.Data
}

func New() *MemStorage {
	return &MemStorage{
		data: make(model.Data),
	}
}

func (s *MemStorage) LoadAll() model.Data {
	return s.data
}

func (s *MemStorage) Load(mtype, mname string) *model.Metric {
	metrics, ok := s.data[mtype]
	if !ok {
		log.Printf("metric type %s doesn't exist", mtype)
		return nil
	}

	mvalue, ok := metrics[mname]
	if !ok {
		log.Printf("metric %s doesn't exist", mvalue)
		return nil
	}

	return &mvalue
}

func (s *MemStorage) StoreCounter(mtype, mname string, mvalue int64) bool {
	counter, ok := s.data[mtype]
	if !ok {
		s.data[mtype] = map[string]model.Metric{
			mname: {Name: mname, Type: mtype, Value: mvalue},
		}
		return true
	}
	val, ok := counter[mname]
	if !ok {
		counter[mname] = model.Metric{Name: mname, Type: mtype, Value: mvalue}
		return true
	}
	v, ok := val.Value.(int64)
	if !ok {
		return false
	}
	v += mvalue
	counter[mname] = model.Metric{Name: mname, Type: mtype, Value: v}
	log.Printf("storage content: %+v\n", s.data)

	return true
}

func (s *MemStorage) StoreGauge(mtype, mname string, mvalue float64) bool {
	gauge, ok := s.data[mtype]
	if !ok {
		s.data[mtype] = map[string]model.Metric{
			mname: {Name: mname, Type: mtype, Value: mvalue},
		}
		return true
	}
	// fix
	_, ok = gauge[mname]
	if !ok {
		gauge[mname] = model.Metric{Name: mname, Type: mtype, Value: mvalue}
		return true
	}
	gauge[mname] = model.Metric{Name: mname, Type: mtype, Value: mvalue}
	log.Printf("storage content: %+v\n", s.data)

	return true
}
