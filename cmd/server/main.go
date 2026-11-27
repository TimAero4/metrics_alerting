package main

import (
	"net/http"
	"os"

	"metrics_alerting/internal/handler"
	"metrics_alerting/internal/middleware" // Добавляем импорт middleware
	"metrics_alerting/internal/repository"
	"metrics_alerting/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap" // Добавляем импорт zap
)

var _ service.Storage = (*repository.MemStorage)(nil)

func main() {
	// Инициализируем zap логгер
	logger, err := zap.NewProduction()
	if err != nil {
		os.Stderr.WriteString("failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer logger.Sync() // Сбрасываем буферы логгера перед выходом

	cfg := ParseConfig() // Теперь функция возвращает структуру Config

	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewHandler(svc)

	r := chi.NewRouter()

	// Подключаем наше middleware для логирования всех запросов
	r.Use(middleware.Logger(logger))

	r.Post("/update/{type}/{name}/{value}", h.UpdateHandler())
	r.Get("/value/{type}/{name}", h.ValueHandler())
	r.Get("/", h.ListHandler())

	logger.Info("Server started", zap.String("address", cfg.ServerAddress))

	// Запускаем сервер
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		logger.Fatal("Server mistake", zap.Error(err))
	}
}
