package service

import (
	"metrics_alerting/internal/model"
)

// Storage interface defined near the consumer
type Storage interface {
	Update(metricType model.MetricType, metricName string, metricValue string) error
	Get(metricType model.MetricType, metricName string) (string, bool)
	GetAll() (gauges map[string]float64, counters map[string]int64)
}

type MetricService struct {
	store Storage
}

func NewMetricService(store Storage) *MetricService {
	return &MetricService{store: store}
}

func (s *MetricService) ProcessUpdate(metricType model.MetricType, name, value string) error {
	return s.store.Update(metricType, name, value)
}

func (s *MetricService) GetMetric(metricType model.MetricType, name string) (string, bool) {
	return s.store.Get(metricType, name)
}

func (s *MetricService) GetAllMetrics() (map[string]float64, map[string]int64) {
	return s.store.GetAll()
}
