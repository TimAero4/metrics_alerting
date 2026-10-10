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
		// Парсим URL: /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		// Разделяем на 5 частей: "", "update", "<ТИП>", "<ИМЯ>", "<ЗНАЧЕНИЕ>"
		parts := strings.Split(r.URL.Path, "/")

		if len(parts) != 5 {
			// Если частей меньше 5, значит отсутствует имя или значение.
			// По требованию: нет имени метрики -> 404
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		metricTypeStr := parts[2]
		metricName := parts[3]
		metricValue := parts[4]

		// Требование: запрос без имени метрики возвращает 404
		if metricName == "" {
			http.Error(w, "Имя метрики обязательно", http.StatusNotFound)
			return
		}

		metricType := MetricType(metricTypeStr)

		// Валидация типа метрики (gauge или counter)
		if metricType != Gauge && metricType != Counter {
			http.Error(w, "Некорректный тип метрики", http.StatusBadRequest)
			return
		}

		// Попытка обновить хранилище. Если формат значения некорректен, возвращаем 400.
		err := store.Update(metricType, metricName, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Успех: возвращаем HTTP 200 OK
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
	mux.HandleFunc("POST /update/", metricsHandler(store))

	//Запускаем http сервер на 8080
	fmt.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server mistake: %v\n", err)
	}
}
