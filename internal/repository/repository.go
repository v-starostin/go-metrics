package repository

import (
	"log"
)

type MemStorage struct {
	data map[string]map[string]any
}

func New() *MemStorage {
	return &MemStorage{
		data: make(map[string]map[string]any),
	}
}

func (s *MemStorage) StoreCounter(mtype, mname string, mvalue int64) bool {
	counter, ok := s.data[mtype]
	if !ok {
		s.data[mtype] = map[string]any{mname: mvalue}
		return true
	}
	val, ok := counter[mname]
	if !ok {
		counter[mname] = mvalue
		return true
	}
	v, ok := val.(int64)
	if !ok {
		return false
	}
	v += mvalue
	counter[mname] = v
	log.Printf("storage content: %+v\n", s.data)

	return true
}

func (s *MemStorage) StoreGauge(mtype, mname string, mvalue float64) bool {
	gauge, ok := s.data[mtype]
	if !ok {
		s.data[mtype] = map[string]any{mname: mvalue}
		return true
	}
	_, ok = gauge[mname]
	if !ok {
		gauge[mname] = mvalue
		return true
	}
	gauge[mname] = mvalue
	log.Printf("storage content: %+v\n", s.data)

	return true
}
