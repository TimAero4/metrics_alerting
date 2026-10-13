package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Storage interface {
	Update(metricType MetricType, metricName string, metricValue string) error
	Get(metricType MetricType, metricName string) (string, bool)
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

// Update обрабатывает входящее значение метрики в зависимости от её типа.
func (s *MemStorage) Update(metricType MetricType, metricName string, metricValue string) error {
	switch metricType {
	case Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("некорректное значение gauge: %w", err)
		}
		// Новое значение замещает предыдущее
		s.GaugeMetrics[metricName] = val

	case Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректное значение counter: %w", err)
		}
		// Новое значение добавляется к предыдущему
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

func metricsHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Жестко отсекаем любые методы, кроме POST (защита от 403 Forbidden)
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// 2. Парсим URL вручную: /update/<ТИП>/<ИМЯ>/<ЗНАЧЕНИЕ>
		parts := strings.Split(r.URL.Path, "/")

		// При пути /update/gauge/test/100 parts будет: ["", "update", "gauge", "test", "100"] (длина 5)
		if len(parts) != 5 {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		metricTypeStr := parts[2]
		metricName := parts[3]
		metricValue := parts[4]

		if metricName == "" {
			http.Error(w, "Имя метрики обязательно", http.StatusNotFound)
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

func main() {
	// Инициализация хранилища в памяти
	store := NewMemStorage()

	// Создаем маршрутизатор (mux)
	mux := http.NewServeMux()

	// Обработчик metricsHandler для update
	mux.HandleFunc("/update/", metricsHandler(store))

	//Запускаем http сервер на 8080
	fmt.Printf("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server mistake: %v\n", err)
	}
}
