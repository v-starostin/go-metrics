package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"ADDRESS"`
	ReportInterval  int    `env:"REPORT_INTERVAL"`
	PollInterval    int    `env:"POLL_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
	StoreInterval   *int   `env:"STORE_INTERVAL"`
	DatabaseDNS     string `env:"DATABASE_DSN"`
}

func NewAgent() (Config, error) {
	flags := parseAgentFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return Config{}, err
	}

	if config.ServerAddress == "" {
		config.ServerAddress = flags.ServerAddress
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = flags.ReportInterval
	}
	if config.PollInterval == 0 {
		config.PollInterval = flags.PollInterval
	}

	return config, nil
}

func NewServer() (Config, error) {
	flags := parseServerFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return Config{}, err
	}

	if config.ServerAddress == "" {
		config.ServerAddress = flags.ServerAddress
	}
	if config.Restore == nil {
		config.Restore = flags.Restore
	}
	if config.StoreInterval == nil {
		config.StoreInterval = flags.StoreInterval
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = flags.FileStoragePath
	}
	if config.DatabaseDNS == "" {
		config.DatabaseDNS = flags.DatabaseDNS
	}

	return config, nil
}

func parseAgentFlags() Config {
	serverAddress := flag.String("a", "localhost:8080", "HTTP server endpoint address")
	reportInterval := flag.Int("r", 10, "report interval to the server (in seconds)")
	pollInterval := flag.Int("p", 2, "interval to gather metrics (in seconds)")
	flag.Parse()

	return Config{
		ServerAddress:  *serverAddress,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
	}
}

func parseServerFlags() Config {
	serverAddress := flag.String("a", "localhost:8080", "address and port to run server")
	fileStoragePath := flag.String("f", "/tmp/metrics-db.json", "file storage path")
	databaseDSN := flag.String("d", "postgres://name:password@localhost:5432/metrics", "database DSN")
	restore := flag.Bool("r", true, "restore")
	storeInterval := flag.Int("i", 300, "interval")
	flag.Parse()

	return Config{
		ServerAddress:   *serverAddress,
		FileStoragePath: *fileStoragePath,
		DatabaseDNS:     *databaseDSN,
		Restore:         restore,
		StoreInterval:   storeInterval,
	}
}
