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
	Key             string `env:"KEY"`
	RateLimit       int    `env:"RATE_LIMIT"`
	CryptoKey       string `env:"CRYPTO_KEY"`
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
	if config.Key == "" {
		config.Key = flags.Key
	}
	if config.RateLimit == 0 {
		config.RateLimit = flags.RateLimit
	}
	if config.CryptoKey == "" {
		config.CryptoKey = flags.CryptoKey
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
	if config.Key == "" {
		config.Key = flags.Key
	}
	if config.RateLimit == 0 {
		config.RateLimit = flags.RateLimit
	}
	if config.CryptoKey == "" {
		config.CryptoKey = flags.CryptoKey
	}

	if config.DatabaseDNS != "" {
		config.FileStoragePath = ""
		*config.StoreInterval = -1
		*config.Restore = false
	}

	return config, nil
}

func parseAgentFlags() Config {
	serverAddress := flag.String("a", "localhost:8080", "HTTP server endpoint address")
	reportInterval := flag.Int("r", 10, "report interval to the server (in seconds)")
	pollInterval := flag.Int("p", 2, "interval to gather metrics (in seconds)")
	key := flag.String("k", "", "key")
	rateLimit := flag.Int("l", 1, "rate limit")
	cryptoKey := flag.String("crypto-key", "", "Path to the public key")
	flag.Parse()

	return Config{
		ServerAddress:  *serverAddress,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
		Key:            *key,
		RateLimit:      *rateLimit,
		CryptoKey:      *cryptoKey,
	}
}

func parseServerFlags() Config {
	serverAddress := flag.String("a", "localhost:8080", "address and port to run server")
	fileStoragePath := flag.String("f", "/tmp/metrics-db.json", "file storage path")
	databaseDSN := flag.String("d", "", "database DSN")
	restore := flag.Bool("r", true, "restore")
	storeInterval := flag.Int("i", 300, "interval")
	key := flag.String("k", "", "")
	cryptoKey := flag.String("crypto-key", "", "Path to the private key")
	flag.Parse()

	return Config{
		ServerAddress:   *serverAddress,
		FileStoragePath: *fileStoragePath,
		DatabaseDNS:     *databaseDSN,
		Restore:         restore,
		StoreInterval:   storeInterval,
		Key:             *key,
		CryptoKey:       *cryptoKey,
	}
}
