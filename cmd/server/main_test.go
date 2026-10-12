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

// TestMetricsHandler тестирует HTTP хендлеры сервера
func TestMetricsHandler(t *testing.T) {
	store := NewMemStorage()

	// Тестируем POST /update/
	t.Run("POST /update/", func(t *testing.T) {
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
	})

	// Подготовим данные для GET хендлеров
	store.Update(Gauge, "Alloc", "100.5")
	store.Update(Counter, "PollCount", "5")

	// Тестируем GET /value/
	t.Run("GET /value/", func(t *testing.T) {
		tests := []struct {
			name       string
			target     string
			wantStatus int
			wantBody   string
		}{
			{"existing gauge", "/value/gauge/Alloc", http.StatusOK, "100.5"},
			{"existing counter", "/value/counter/PollCount", http.StatusOK, "5"},
			{"unknown metric", "/value/gauge/Unknown", http.StatusNotFound, ""},
			{"unknown type", "/value/unknown/Alloc", http.StatusNotFound, ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, tt.target, nil)
				w := httptest.NewRecorder()

				mux := http.NewServeMux()
				mux.HandleFunc("GET /value/{type}/{name}", valueHandler(store))
				mux.ServeHTTP(w, req)

				if w.Code != tt.wantStatus {
					t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
				}
				if tt.wantBody != "" && w.Body.String() != tt.wantBody {
					t.Errorf("expected body %q, got %q", tt.wantBody, w.Body.String())
				}
			})
		}
	})

	// Тестируем GET /
	t.Run("GET /", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Вызываем хендлер напрямую
		listHandler(store).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Errorf("expected text/html content type")
		}
		// Проверяем, что в HTML есть наши метрики
		body := w.Body.String()
		if !contains(body, "Alloc") || !contains(body, "PollCount") {
			t.Error("HTML body does not contain expected metrics")
		}
	})
}

// вспомогательная функция для проверки подстроки
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
