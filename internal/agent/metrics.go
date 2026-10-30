package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
)

type MetricsStore struct {
	mu      sync.Mutex
	gauge   map[string]float64
	counter map[string]int64
}

func NewMetricStore() *MetricsStore {
	return &MetricsStore{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (s *MetricsStore) GetGauge(name string) (float64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.gauge[name]
	return val, ok
}

func (s *MetricsStore) GetCounter(name string) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.counter[name]
	return val, ok
}

func (s *MetricsStore) UpdateRuntimeMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	s.gauge["Alloc"] = float64(m.Alloc)
	s.gauge["BuckHashSys"] = float64(m.BuckHashSys)
	s.gauge["Frees"] = float64(m.Frees)
	s.gauge["GCCPUFraction"] = m.GCCPUFraction
	s.gauge["GCSys"] = float64(m.GCSys)
	s.gauge["HeapAlloc"] = float64(m.HeapAlloc)
	s.gauge["HeapIdle"] = float64(m.HeapIdle)
	s.gauge["HeapInuse"] = float64(m.HeapInuse)
	s.gauge["HeapObjects"] = float64(m.HeapObjects)
	s.gauge["HeapReleased"] = float64(m.HeapReleased)
	s.gauge["HeapSys"] = float64(m.HeapSys)
	s.gauge["LastGC"] = float64(m.LastGC)
	s.gauge["Lookups"] = float64(m.Lookups)
	s.gauge["MCacheInuse"] = float64(m.MCacheInuse)
	s.gauge["MCacheSys"] = float64(m.MCacheSys)
	s.gauge["MSpanInuse"] = float64(m.MSpanInuse)
	s.gauge["MSpanSys"] = float64(m.MSpanSys)
	s.gauge["Mallocs"] = float64(m.Mallocs)
	s.gauge["NextGC"] = float64(m.NextGC)
	s.gauge["NumForcedGC"] = float64(m.NumForcedGC)
	s.gauge["NumGC"] = float64(m.NumGC)
	s.gauge["OtherSys"] = float64(m.OtherSys)
	s.gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	s.gauge["StackInuse"] = float64(m.StackInuse)
	s.gauge["StackSys"] = float64(m.StackSys)
	s.gauge["Sys"] = float64(m.Sys)
	s.gauge["TotalAlloc"] = float64(m.TotalAlloc)
	s.gauge["RandomValue"] = rand.Float64()

	s.counter["PollCount"]++
}

func (s *MetricsStore) SendMetrics(serverAddress string) {
	s.mu.Lock()
	gaugesToSend := make(map[string]float64)
	for k, v := range s.gauge {
		gaugesToSend[k] = v
	}
	countersToSend := make(map[string]int64)
	for k, v := range s.counter {
		countersToSend[k] = v
	}

	// Сбрасываем счетчики сразу после копирования, под тем же мьютексом
	for k := range s.counter {
		s.counter[k] = 0
	}

	s.mu.Unlock()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	for name, value := range gaugesToSend {
		valueStr := strconv.FormatFloat(value, 'f', -1, 64)
		sendRequest(client, serverAddress, "gauge", name, valueStr)
	}

	for name, value := range countersToSend {
		valueStr := strconv.FormatInt(value, 10)
		sendRequest(client, serverAddress, "counter", name, valueStr)
	}
}

func sendRequest(client *http.Client, serverAddress, metricType, metricName, metricValue string) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddress, metricType, metricName, metricValue)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Printf("Error creating request for %s: %v\n", metricName, err)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending metric %s: %v\n", metricName, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Server returned status %d for metric %s\n", resp.StatusCode, metricName)
	}
}
