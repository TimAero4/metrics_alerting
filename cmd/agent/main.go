package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"metrics_alerting/internal/agent"
)

func main() {
	cfg := ParseConfig() // Теперь используем единую функцию

	store := agent.NewMetricStore()

	log.Printf("Agent started. Poll every %v, report every %v to %s\n", cfg.PollInterval, cfg.ReportInterval, cfg.ServerAddress)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	go func() {
		ticker := time.NewTicker(cfg.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				store.UpdateRuntimeMetrics()
			case <-ctx.Done():
				return
			}
		}
	}()

	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer reportTicker.Stop()
	for {
		select {
		case <-reportTicker.C:
			store.SendMetrics(cfg.ServerAddress)
		case <-ctx.Done():
			return
		}
	}
}
