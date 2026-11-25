package repository

import (
	"fmt"
	"strconv"
	"sync"

	"metrics_alerting/internal/model"
)

type MemStorage struct {
	mu             sync.Mutex
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}
}

func (s *MemStorage) Update(metricType model.MetricType, metricName string, metricValue string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch metricType {
	case model.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("incorrect gauge value: %w", err)
		}
		s.gaugeMetrics[metricName] = val
	case model.Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("incorrect counter value: %w", err)
		}
		s.counterMetrics[metricName] += val
	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
	return nil
}

func (s *MemStorage) Get(metricType model.MetricType, metricName string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch metricType {
	case model.Gauge:
		val, ok := s.gaugeMetrics[metricName]
		if !ok {
			return "", false
		}
		return strconv.FormatFloat(val, 'f', -1, 64), true
	case model.Counter:
		val, ok := s.counterMetrics[metricName]
		if !ok {
			return "", false
		}
		return strconv.FormatInt(val, 10), true
	default:
		return "", false
	}
}

func (s *MemStorage) GetAll() (gauges map[string]float64, counters map[string]int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gauges = make(map[string]float64)
	for k, v := range s.gaugeMetrics {
		gauges[k] = v
	}
	counters = make(map[string]int64)
	for k, v := range s.counterMetrics {
		counters[k] = v
	}
	return
}
