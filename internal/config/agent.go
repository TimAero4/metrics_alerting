package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

type AgentConfig struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func ParseAgentConfig() AgentConfig {
	// Создаем локальный FlagSet вместо использования глобального flag.CommandLine
	// Это позволяет безопасно вызывать функцию несколько раз в тестах
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	addrFlag := fs.String("a", "", "Address of the HTTP server endpoint")
	reportIntervalFlag := fs.Int("r", 10, "Report interval in seconds")
	pollIntervalFlag := fs.Int("p", 2, "Poll interval in seconds")

	// Парсим аргументы командной строки через наш локальный FlagSet
	// Ошибки игнорируются, чтобы поддержать работу флагов вроде -test.v в тестах Go
	_ = fs.Parse(os.Args[1:])

	// --- Server Address ---
	addr := ""
	// Исправлено на ADDRESS согласно заданию
	if envAddr, ok := os.LookupEnv("ADDRESS"); ok {
		addr = envAddr
	} else if *addrFlag != "" {
		addr = *addrFlag
	} else {
		addr = "http://localhost:8080"
	}

	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	// --- Report Interval ---
	reportInterval := time.Duration(*reportIntervalFlag) * time.Second
	if envReport, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		if secs, err := strconv.Atoi(envReport); err == nil {
			reportInterval = time.Duration(secs) * time.Second
		}
	}

	// --- Poll Interval ---
	pollInterval := time.Duration(*pollIntervalFlag) * time.Second
	if envPoll, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		if secs, err := strconv.Atoi(envPoll); err == nil {
			pollInterval = time.Duration(secs) * time.Second
		}
	}

	return AgentConfig{
		ServerAddress:  addr,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
	}
}
