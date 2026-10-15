package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMemStorage_Update(t *testing.T) {
	store := NewMemStorage()

	err := store.Update(Gauge, "testGauge", "123.45")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := store.Get(Gauge, "testGauge")
	if !ok || val != "123.45" {
		t.Errorf("expected 123.45, got %s", val)
	}

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

	err = store.Update(Gauge, "badGauge", "not-a-float")
	if err == nil {
		t.Error("expected error for bad gauge value, got nil")
	}
}

func TestUpdateHandler(t *testing.T) {
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
			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", updateHandler(store))

			req := httptest.NewRequest(tt.method, tt.target, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestValueHandler(t *testing.T) {
	store := NewMemStorage()
	store.Update(Gauge, "TestGauge", "100.5")
	store.Update(Counter, "TestCounter", "5")

	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", valueHandler(store))

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
			}
			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Errorf("expected body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}
