package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

// Storage интерфейс определяет методы для взаимодействия с хранилищем.
type Storage interface {
	Update(metricType MetricType, metricName string, metricValue string) error
	Get(metricType MetricType, metricName string) (string, bool)
	GetAll() (gauges map[string]float64, counters map[string]int64)
}

type MemStorage struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (s *MemStorage) Update(metricType MetricType, metricName string, metricValue string) error {
	switch metricType {
	case Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("некорректное значение gauge: %w", err)
		}
		s.GaugeMetrics[metricName] = val
	case Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректное значение counter: %w", err)
		}
		s.CounterMetrics[metricName] += val
	default:
		return fmt.Errorf("неизвестный тип метрики: %s", metricType)
	}
	return nil
}

func (s *MemStorage) Get(metricType MetricType, metricName string) (string, bool) {
	switch metricType {
	case Gauge:
		val, ok := s.GaugeMetrics[metricName]
		if !ok {
			return "", false
		}
		return strconv.FormatFloat(val, 'f', -1, 64), true
	case Counter:
		val, ok := s.CounterMetrics[metricName]
		if !ok {
			return "", false
		}
		return strconv.FormatInt(val, 10), true
	default:
		return "", false
	}
}

// GetAll возвращает копии всех метрик для формирования HTML страницы.
func (s *MemStorage) GetAll() (gauges map[string]float64, counters map[string]int64) {
	gauges = make(map[string]float64)
	for k, v := range s.GaugeMetrics {
		gauges[k] = v
	}
	counters = make(map[string]int64)
	for k, v := range s.CounterMetrics {
		counters[k] = v
	}
	return
}

// updateHandler обрабатывает POST /update/{type}/{name}/{value}
func updateHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricTypeStr := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		if metricName == "" || metricValue == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		metricType := MetricType(metricTypeStr)
		if metricType != Gauge && metricType != Counter {
			http.Error(w, "Некорректный тип метрики", http.StatusBadRequest)
			return
		}

		err := store.Update(metricType, metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// valueHandler обрабатывает GET /value/{type}/{name}
func valueHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricTypeStr := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		if metricName == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		metricType := MetricType(metricTypeStr)
		val, ok := store.Get(metricType, metricName)
		if !ok {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(val))
	}
}

// listHandler обрабатывает GET / и возвращает HTML страницу со всеми метриками
func listHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gauges, counters := store.GetAll()

		var sb strings.Builder
		sb.WriteString("<html><body>")
		sb.WriteString("<h1>Gauge Metrics</h1><ul>")
		for name, value := range gauges {
			sb.WriteString(fmt.Sprintf("<li>%s: %s</li>", name, strconv.FormatFloat(value, 'f', -1, 64)))
		}
		sb.WriteString("</ul>")

		sb.WriteString("<h1>Counter Metrics</h1><ul>")
		for name, value := range counters {
			sb.WriteString(fmt.Sprintf("<li>%s: %s</li>", name, strconv.FormatInt(value, 10)))
		}
		sb.WriteString("</ul>")

		sb.WriteString("</body></html>")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sb.String()))
	}
}

func main() {
	// 1. Парсим флаги и окружения
	ParseFlags()
	serverAddr := GetServerAddress()

	store := NewMemStorage()

	// Используем роутер chi
	r := chi.NewRouter()

	// Регистрируем маршруты
	r.Post("/update/{type}/{name}/{value}", updateHandler(store))
	r.Get("/value/{type}/{name}", valueHandler(store))
	r.Get("/", listHandler(store))

	fmt.Printf("Сервер запущен %v\n", serverAddr)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		fmt.Printf("Server mistake: %v\n", err)
	}
}
