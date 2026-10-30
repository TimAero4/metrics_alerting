package main

import (
	"flag"
	"os"
	"sync"
	"testing"
	"time"

	"metrics_alerting/internal/agent"
)

// TestUpdateRuntimeMetrics тестирует логику сбора метрик через внутренний пакет
func TestUpdateRuntimeMetrics(t *testing.T) {
	store := agent.NewMetricStore()

	store.UpdateRuntimeMetrics()

	// Проверяем gauge метрики через геттеры
	if _, ok := store.GetGauge("Alloc"); !ok {
		t.Error("Expected 'Alloc' gauge metric to be set")
		return // Добавляем return после t.Error, как просил ревьюер
	}
	if _, ok := store.GetGauge("RandomValue"); !ok {
		t.Error("Expected 'RandomValue' gauge metric to be set")
		return
	}

	// Проверяем счетчик PollCount
	if val, _ := store.GetCounter("PollCount"); val != 1 {
		t.Errorf("Expected PollCount to be 1, got %d", val)
		return
	}

	store.UpdateRuntimeMetrics()

	if val, _ := store.GetCounter("PollCount"); val != 2 {
		t.Errorf("Expected PollCount to be 2 after second update, got %d", val)
	}
}

// TestConcurrentUpdate тестирует потокобезопасность хранилища агента
func TestConcurrentUpdate(t *testing.T) {
	store := agent.NewMetricStore()
	var wg sync.WaitGroup

	// Запустим 100 горутин, одновременно обновляющих метрики
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.UpdateRuntimeMetrics()
		}()
	}

	wg.Wait() // Ждем завершения всех горутин

	if val, _ := store.GetCounter("PollCount"); val != 100 {
		t.Errorf("Expected PollCount to be 100 after concurrent updates, got %d", val)
	}
}

// TestParseConfig тестирует единую функцию конфигурации
func TestParseConfig(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name     string
		args     []string
		wantAddr string
		wantPoll time.Duration
	}{
		{"default values", []string{"cmd"}, "http://localhost:8080", 2 * time.Second},
		{"custom flags", []string{"cmd", "-a=http://localhost:9090", "-p=5", "-r=30"}, "http://localhost:9090", 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем состояние флагов для каждого теста
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			os.Args = tt.args

			cfg := ParseConfig()

			if cfg.ServerAddress != tt.wantAddr {
				t.Errorf("ServerAddress = %v, want %v", cfg.ServerAddress, tt.wantAddr)
				return
			}
			if cfg.PollInterval != tt.wantPoll {
				t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, tt.wantPoll)
			}
		})
	}
}
