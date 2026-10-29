package main

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	ServerAddress string
}

// ParseConfig собирает конфигурацию с учетом приоритета: ENV -> Flag -> Default
func ParseConfig() Config {
	addrFlag := flag.String("a", "", "Address of the HTTP server endpoint")
	flag.Parse()

	addr := ""
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		addr = envAddr
	} else if *addrFlag != "" {
		addr = *addrFlag
	} else {
		addr = ":8080" // Дефолт для сервера
	}

	// Автотесты могут передать адрес в формате localhost:port.
	// http.ListenAndServe требует формат :port или host:port.
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	return Config{
		ServerAddress: addr,
	}
}
