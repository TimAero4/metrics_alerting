package main

import (
	"flag"
	"os"
	"sync"
	"testing"
	"time"
)

// TestUpdateRuntimeMetrics тестирует логику сбора метрик агентом
func TestUpdateRuntimeMetrics(t *testing.T) {
	store := NewMetricStore()

	// Первый сбор
	store.UpdateRuntimeMetrics()

	// Проверяем, что gauge метрики заполнились
	if _, ok := store.Gauge["Alloc"]; !ok {
		t.Error("Expected 'Alloc' gauge metric to be set")
	}
	if _, ok := store.Gauge["RandomValue"]; !ok {
		t.Error("Expected 'RandomValue' gauge metric to be set")
	}

	// Проверяем, что счетчик PollCount увеличился до 1
	if store.Counter["PollCount"] != 1 {
		t.Errorf("Expected PollCount to be 1, got %d", store.Counter["PollCount"])
	}

	// Второй сбор
	store.UpdateRuntimeMetrics()

	// Проверяем, что счетчик увеличился
	if store.Counter["PollCount"] != 2 {
		t.Errorf("Expected PollCount to be 2 after second update, got %d", store.Counter["PollCount"])
	}
}

// TestConcurrentUpdate тестирует потокобезопасность хранилища агента
func TestConcurrentUpdate(t *testing.T) {
	store := NewMetricStore()
	wg := sync.WaitGroup{}

	// Запустим 100 горутин, одновременно обновляющих метрики
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.UpdateRuntimeMetrics()
		}()
	}

	wg.Wait()

	// Если мьютекс работает неправильно, программа упадет с fatal error: concurrent map writes
	// Проверяем, что счетчик в итоге равен ровно 100
	if store.Counter["PollCount"] != 100 {
		t.Errorf("Expected PollCount to be 100 after concurrent updates, got %d", store.Counter["PollCount"])
	}
}

func TestGetServerAddress(t *testing.T) {
	origArgs := os.Args
	origEnv := os.Getenv("SERVER_ADDRESS")
	defer func() {
		os.Args = origArgs
		os.Setenv("SERVER_ADDRESS", origEnv)
	}()

	tests := []struct {
		name    string
		args    []string
		envAddr string
		want    string
	}{
		{"default value", []string{"cmd"}, "", "http://localhost:8080"},
		{"flag value", []string{"cmd", "-a=http://localhost:9090"}, "", "http://localhost:9090"},
		{"env value", []string{"cmd"}, "http://localhost:7070", "http://localhost:7070"},
		{"env overrides flag", []string{"cmd", "-a=http://localhost:9090"}, "http://localhost:7070", "http://localhost:7070"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем состояние флагов для каждого теста
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			// Переопределяем флаги, так как они привязаны к новому FlagSet
			addrFlag = flag.String("a", "", "Address of the HTTP server endpoint")
			reportIntervalFlag = flag.Int("r", 10, "Report interval in seconds")
			pollIntervalFlag = flag.Int("p", 2, "Poll interval in seconds")

			os.Args = tt.args
			os.Setenv("SERVER_ADDRESS", tt.envAddr)
			ParseFlags()

			got := GetServerAddress()
			if got != tt.want {
				t.Errorf("GetServerAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPollInterval(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name string
		args []string
		want time.Duration
	}{
		{"default", []string{"cmd"}, 2 * time.Second},
		{"custom", []string{"cmd", "-p=5"}, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			addrFlag = flag.String("a", "", "Address of the HTTP server endpoint")
			reportIntervalFlag = flag.Int("r", 10, "Report interval in seconds")
			pollIntervalFlag = flag.Int("p", 2, "Poll interval in seconds")

			os.Args = tt.args
			ParseFlags()

			got := GetPollInterval()
			if got != tt.want {
				t.Errorf("GetPollInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetReportInterval(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name string
		args []string
		want time.Duration
	}{
		{"default", []string{"cmd"}, 10 * time.Second},
		{"custom", []string{"cmd", "-r=30"}, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			addrFlag = flag.String("a", "", "Address of the HTTP server endpoint")
			reportIntervalFlag = flag.Int("r", 10, "Report interval in seconds")
			pollIntervalFlag = flag.Int("p", 2, "Poll interval in seconds")

			os.Args = tt.args
			ParseFlags()

			got := GetReportInterval()
			if got != tt.want {
				t.Errorf("GetReportInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
