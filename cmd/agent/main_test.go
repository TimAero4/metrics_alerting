package main

import (
	"sync"
	"testing"
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
