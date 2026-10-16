package main

import (
	"flag"
	"os"
	"strings"
	"time"
)

var (
	addrFlag           = flag.String("a", "", "Address of the HTTP server endpoint")
	reportIntervalFlag = flag.Int("r", 10, "Report interval in seconds")
	pollIntervalFlag   = flag.Int("p", 2, "Poll interval in seconds")
)

func ParseFlags() {
	flag.Parse()
}

// GetServerAddress применяет приоритет: ENV -> Flag -> Default
func GetServerAddress() string {
	addr := ""
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		addr = envAddr
	} else if *addrFlag != "" {
		addr = *addrFlag
	} else {
		addr = "http://localhost:8080" // Дефолт для агента
	}

	// Если тесты передают адрес без http:// (например, localhost:8080).
	// Если протокол не указан, добавляем его принудительно.
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	return addr
}

func GetReportInterval() time.Duration {
	return time.Duration(*reportIntervalFlag) * time.Second
}

func GetPollInterval() time.Duration {
	return time.Duration(*pollIntervalFlag) * time.Second
}
