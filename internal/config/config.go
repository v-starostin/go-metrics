package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func NewAgent() *Config {
	srvAddr, reportInt, pollInt := parseAgentFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	if config.ServerAddress == "" {
		config.ServerAddress = srvAddr
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = reportInt
	}
	if config.PollInterval == 0 {
		config.PollInterval = pollInt
	}

	return &config
}

func NewServer() *Config {
	srvAddr := parseServerFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	if config.ServerAddress == "" {
		config.ServerAddress = srvAddr
	}

	return &config
}

func parseAgentFlags() (string, int, int) {
	serverAddress := flag.String("a", "localhost:8080", "HTTP server endpoint address")
	reportInterval := flag.Int("r", 10, "report interval to the server (in seconds)")
	pollInterval := flag.Int("p", 2, "interval to gather metrics (in seconds)")
	flag.Parse()

	return *serverAddress, *reportInterval, *pollInterval
}

func parseServerFlags() string {
	serverAddress := flag.String("a", "localhost:8080", "address and port to run server")
	flag.Parse()

	return *serverAddress
}
