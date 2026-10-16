package main

import (
	"flag"
	"os"
)

var addrFlag = flag.String("a", "", "Address of the HTTP server endpoint")

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
	return ":8080" // Дефолт для сервера
}
