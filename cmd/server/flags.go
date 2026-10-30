package main

import "metrics_alerting/internal/config"

// ParseConfig делегирует разбор конфигурации пакету internal/config.
// Это позволяет тестировать логику конфигурации изолированно от точки входа (main).
func ParseConfig() config.ServerConfig {
	return config.ParseServerConfig()
}
