package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// MetricStore contain all metrics and protect them by mutex
type MetricsStore struct {
	mu      sync.Mutex
	Gauge   map[string]float64
	Counter map[string]int64
}

// NewMetricStore make new metric store
func NewMetricStore() *MetricsStore {
	return &MetricsStore{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

// UpdateRuntimeMetrics
func (s *MetricsStore) UpdateRuntimeMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Update gauge metrics
	s.Gauge["Alloc"] = float64(m.Alloc)
	s.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
	s.Gauge["Frees"] = float64(m.Frees)
	s.Gauge["GCCPUFraction"] = m.GCCPUFraction
	s.Gauge["GCSys"] = float64(m.GCSys)
	s.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
	s.Gauge["HeapIdle"] = float64(m.HeapIdle)
	s.Gauge["HeapInuse"] = float64(m.HeapInuse)
	s.Gauge["HeapObjects"] = float64(m.HeapObjects)
	s.Gauge["HeapReleased"] = float64(m.HeapReleased)
	s.Gauge["HeapSys"] = float64(m.HeapSys)
	s.Gauge["LastGC"] = float64(m.LastGC)
	s.Gauge["Lookups"] = float64(m.Lookups)
	s.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
	s.Gauge["MCacheSys"] = float64(m.MCacheSys)
	s.Gauge["MSpanInuse"] = float64(m.MSpanInuse)
	s.Gauge["MSpanSys"] = float64(m.MSpanSys)
	s.Gauge["Mallocs"] = float64(m.Mallocs)
	s.Gauge["NextGC"] = float64(m.NextGC)
	s.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
	s.Gauge["NumGC"] = float64(m.NumGC)
	s.Gauge["OtherSys"] = float64(m.OtherSys)
	s.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	s.Gauge["StackInuse"] = float64(m.StackInuse)
	s.Gauge["StackSys"] = float64(m.StackSys)
	s.Gauge["Sys"] = float64(m.Sys)
	s.Gauge["TotalAlloc"] = float64(m.TotalAlloc)

	s.Gauge["RandomValue"] = rand.Float64()

	s.Counter["PollCount"]++
}

func (s *MetricsStore) SendMetrics(serverAddress string) {
	s.mu.Lock()
	// Копируем данные, чтобы не держать мьютекс заблокированным во время сетевых запросов
	gaugesToSend := make(map[string]float64)
	for k, v := range s.Gauge {
		gaugesToSend[k] = v
	}
	countersToSend := make(map[string]int64)
	for k, v := range s.Counter {
		countersToSend[k] = v
	}
	s.mu.Unlock()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	// Send gauge
	for name, value := range gaugesToSend {
		// Используем FormatFloat с флагом 'f' и точностью -1,
		// чтобы избежать экспоненциальной записи (например, 1.23e+04),
		// которую сервер может не распарсить как float64
		valueStr := strconv.FormatFloat(value, 'f', -1, 64)
		sendRequest(client, serverAddress, "gauge", name, valueStr)
	}

	// Send counter
	for name, value := range countersToSend {
		valueStr := strconv.FormatInt(value, 10)
		sendRequest(client, serverAddress, "counter", name, valueStr)
	}
}

func sendRequest(client *http.Client, serverAddress, metricType, metricName, metricValue string) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddress, metricType, metricName, metricValue)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Printf("Ошибка создания запроса для %s: %v\n", metricName, err)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка отправки метрики %s: %v\n", metricName, err)
		return
	}
	defer resp.Body.Close()

	// Если сервер вернул не 200, логируем это
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Сервер вернул статус %d для метрики %s\n", resp.StatusCode, metricName)
	}
}

func main() {
	// Настройки агента
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second
	serverAddress := "http://localhost:8080"

	store := NewMetricStore()

	fmt.Printf("Агент запущен. Сбор каждые %v, отправка каждые %v на %s\n", pollInterval, reportInterval, serverAddress)

	// goroutine for metrics
	go func() {
		for {
			store.UpdateRuntimeMetrics()
			time.Sleep(pollInterval)
		}
	}()

	// 1 seconds latency
	time.Sleep(1 * time.Second)

	// sending metrics to server (http)
	for {
		store.SendMetrics(serverAddress)
		time.Sleep(reportInterval)
	}
}
