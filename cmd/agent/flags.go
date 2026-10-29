package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

// ParseConfig собирает конфигурацию агента
func ParseConfig() Config {
	addrFlag := flag.String("a", "", "Address of the HTTP server endpoint")
	reportIntervalFlag := flag.Int("r", 10, "Report interval in seconds")
	pollIntervalFlag := flag.Int("p", 2, "Poll interval in seconds")

	flag.Parse()

	// --- Server Address ---
	addr := ""
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		addr = envAddr
	} else if *addrFlag != "" {
		addr = *addrFlag
	} else {
		addr = "http://localhost:8080" // Дефолт для агента
	}

	// Если протокол не указан, добавляем его принудительно.
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	// --- Report Interval ---
	reportInterval := time.Duration(*reportIntervalFlag) * time.Second
	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		if val, err := time.ParseDuration(envReport); err == nil {
			reportInterval = val
		} else if secs, err := strconv.Atoi(envReport); err == nil {
			reportInterval = time.Duration(secs) * time.Second
		}
	}

	// --- Poll Interval ---
	pollInterval := time.Duration(*pollIntervalFlag) * time.Second
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		if val, err := time.ParseDuration(envPoll); err == nil {
			pollInterval = val
		} else if secs, err := strconv.Atoi(envPoll); err == nil {
			pollInterval = time.Duration(secs) * time.Second
		}
	}

	return Config{
		ServerAddress:  addr,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
	}
}
