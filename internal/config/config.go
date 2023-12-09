package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func NewAgent() *Config {
	srvAddr, reportInt, pollInt := ParseAgentFlags()

	agentConfig := &Config{}
	if err := env.Parse(agentConfig); err != nil {
		log.Fatal(err)
	}

	if agentConfig.ServerAddress == "" {
		agentConfig.ServerAddress = srvAddr
	}
	if agentConfig.ReportInterval == 0 {
		agentConfig.ReportInterval = reportInt
	}
	if agentConfig.PollInterval == 0 {
		agentConfig.PollInterval = pollInt
	}

	return agentConfig
}

func NewServer() *Config {
	srvAddr := ParseServerFlags()

	config := &Config{}
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	if config.ServerAddress == "" {
		config.ServerAddress = srvAddr
	}

	return config
}

func ParseAgentFlags() (string, int, int) {
	serverAddress := flag.String("a", "localhost:8080", "HTTP server endpoint address")
	reportInterval := flag.Int("r", 10, "report interval to the server (in seconds)")
	pollInterval := flag.Int("p", 2, "interval to gather metrics (in seconds)")
	flag.Parse()

	return *serverAddress, *reportInterval, *pollInterval
}

func ParseServerFlags() string {
	serverAddress := flag.String("a", "localhost:8080", "address and port to run server")
	flag.Parse()

	return *serverAddress
}
