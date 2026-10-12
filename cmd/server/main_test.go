package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestMemStorage_Update тестирует логику обновления метрик в хранилище
func TestMemStorage_Update(t *testing.T) {
	store := NewMemStorage()

	// Тест gauge: новое значение должно перезаписывать старое
	err := store.Update(Gauge, "testGauge", "123.45")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := store.Get(Gauge, "testGauge")
	if !ok || val != "123.45" {
		t.Errorf("expected 123.45, got %s", val)
	}

	err = store.Update(Gauge, "testGauge", "999.99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok = store.Get(Gauge, "testGauge")
	if !ok || val != "999.99" {
		t.Errorf("expected 999.99 after override, got %s", val)
	}

	// Тест counter: новое значение должно прибавляться к старому
	err = store.Update(Counter, "testCounter", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok = store.Get(Counter, "testCounter")
	if !ok || val != "10" {
		t.Errorf("expected 10, got %s", val)
	}

	err = store.Update(Counter, "testCounter", "5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok = store.Get(Counter, "testCounter")
	if !ok || val != "15" {
		t.Errorf("expected 15 (10+5), got %s", val)
	}

	// Тест ошибок парсинга
	err = store.Update(Gauge, "badGauge", "not-a-float")
	if err == nil {
		t.Error("expected error for bad gauge value, got nil")
	}
	err = store.Update(Counter, "badCounter", "not-an-int")
	if err == nil {
		t.Error("expected error for bad counter value, got nil")
	}
}

// TestMetricsHandler тестирует HTTP хендлер обновления метрик
func TestMetricsHandler(t *testing.T) {
	store := NewMemStorage()

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
			req := httptest.NewRequest(tt.method, tt.target, nil)
			w := httptest.NewRecorder()

			// Используем стандартный мукс для тестирования, как в main
			mux := http.NewServeMux()
			mux.HandleFunc("POST /update/", metricsHandler(store))
			mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
