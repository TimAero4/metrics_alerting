package main

import (
	"flag"
	"os"
	"strings"
)

var addrFlag = flag.String("a", "", "Address of the HTTP server endpoint")

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
		addr = ":8080" // Дефолт для сервера
	}

	// Автотесты могут передать адрес в формате localhost:port.
	// http.ListenAndServe требует формат :port или host:port.
	// Если нет двоеточия, добавляем его в начало.
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	return addr
}
