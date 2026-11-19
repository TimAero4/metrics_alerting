package config

import (
	"flag"
	"os"
	"strings"
)

type ServerConfig struct {
	ServerAddress string
}

func ParseServerConfig() ServerConfig {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	addrFlag := fs.String("a", ":8080", "Address of the HTTP server endpoint")
	_ = fs.Parse(os.Args[1:])

	// 1. Приоритет переменной окружения ADDRESS
	addr := ""
	if envAddr, ok := os.LookupEnv("ADDRESS"); ok {
		addr = envAddr
	} else if *addrFlag != "" {
		addr = *addrFlag
	} else {
		addr = ":8080"
	}

	// 2. Сервер должен слушать адрес без префикса http://
	// Если префикс пришел из окружения, аккуратно его убираем
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")

	return ServerConfig{
		ServerAddress: addr,
	}
}
