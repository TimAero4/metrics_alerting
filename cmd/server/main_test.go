package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"metrics_alerting/internal/handler"
	"metrics_alerting/internal/model"
	"metrics_alerting/internal/repository"
	"metrics_alerting/internal/service"

	"github.com/go-chi/chi/v5"
)

// TestParseConfig тестирует конфигурацию сервера
func TestParseConfig(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"default value", []string{"cmd"}, ":8080"},
		{"flag value", []string{"cmd", "-a=:9090"}, ":9090"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			os.Args = tt.args

			cfg := ParseConfig()
			if cfg.ServerAddress != tt.want {
				t.Errorf("ServerAddress = %v, want %v", cfg.ServerAddress, tt.want)
				return
			}
		})
	}
}

// TestMemStorage_Update тестирует хранилище через внутренний пакет
func TestMemStorage_Update(t *testing.T) {
	store := repository.NewMemStorage()

	err := store.Update(model.Gauge, "testGauge", "123.45")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := store.Get(model.Gauge, "testGauge")
	if !ok || val != "123.45" {
		t.Errorf("expected 123.45, got %s", val)
		return
	}

	err = store.Update(model.Counter, "testCounter", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok = store.Get(model.Counter, "testCounter")
	if !ok || val != "10" {
		t.Errorf("expected 10, got %s", val)
		return
	}

	err = store.Update(model.Counter, "testCounter", "5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok = store.Get(model.Counter, "testCounter")
	if !ok || val != "15" {
		t.Errorf("expected 15 (10+5), got %s", val)
		return
	}

	err = store.Update(model.Gauge, "badGauge", "not-a-float")
	if err == nil {
		t.Error("expected error for bad gauge value, got nil")
	}
}

// TestUpdateHandler тестирует хендлеры через N-tier архитектуру
func TestUpdateHandler(t *testing.T) {
	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewHandler(svc)

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
	}{
		{"valid gauge", http.MethodPost, "/update/gauge/testG/123.45", http.StatusOK},
		{"valid counter", http.MethodPost, "/update/counter/testC/10", http.StatusOK},
		{"invalid type", http.MethodPost, "/update/unknown/testM/10", http.StatusBadRequest},
		{"invalid value", http.MethodPost, "/update/gauge/testM/abc", http.StatusBadRequest},
		{"empty name", http.MethodPost, "/update/gauge//123", http.StatusNotFound},
		{"wrong method", http.MethodGet, "/update/gauge/testG/123", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", h.UpdateHandler())

			req := httptest.NewRequest(tt.method, tt.target, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// TestValueHandler тестирует получение значений
func TestValueHandler(t *testing.T) {
	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewHandler(svc)

	// Подготовка данных
	store.Update(model.Gauge, "TestGauge", "100.5")
	store.Update(model.Counter, "TestCounter", "5")

	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", h.ValueHandler())

	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantBody   string
	}{
		{"existing gauge", "/value/gauge/TestGauge", http.StatusOK, "100.5"},
		{"existing counter", "/value/counter/TestCounter", http.StatusOK, "5"},
		{"unknown metric", "/value/gauge/Unknown", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
				return
			}
			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Errorf("expected body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}
