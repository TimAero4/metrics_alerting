package main

import (
	"log"
	"net/http"

	"metrics_alerting/internal/handler"
	"metrics_alerting/internal/repository"
	"metrics_alerting/internal/service"

	"github.com/go-chi/chi/v5"
)

var _ service.Storage = (*repository.MemStorage)(nil)

func main() {
	cfg := ParseConfig() // Теперь функция возвращает структуру Config

	store := repository.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewHandler(svc)

	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", h.UpdateHandler())
	r.Get("/value/{type}/{name}", h.ValueHandler())
	r.Get("/", h.ListHandler())

	log.Printf("Server started at %v\n", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatalf("Server mistake: %v\n", err)
	}
}
