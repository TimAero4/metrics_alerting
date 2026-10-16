package main

import (
	"flag"
	"os"
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
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		return envAddr
	}
	if *addrFlag != "" {
		return *addrFlag
	}
	return "http://localhost:8080" // Дефолт для агента
}

func GetReportInterval() time.Duration {
	return time.Duration(*reportIntervalFlag) * time.Second
}

func GetPollInterval() time.Duration {
	return time.Duration(*pollIntervalFlag) * time.Second
}
