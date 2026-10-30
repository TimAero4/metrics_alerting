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
	addrFlag := flag.String("a", "", "Address of the HTTP server endpoint")
	flag.Parse()

	addr := ""
	if envAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		addr = envAddr
	} else if *addrFlag != "" {
		addr = *addrFlag
	} else {
		addr = ":8080"
	}

	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	return ServerConfig{
		ServerAddress: addr,
	}
}
