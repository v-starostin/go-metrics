package config

import (
	"encoding/json"
	"flag"
	"os"

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
	JSONConfigPath  string `env:"CONFIG"`
}

func NewAgent() (Config, error) {
	flags := parseAgentFlags()

	if flags.JSONConfigPath != "" {
		jsonConfig, err := parseJSONConfig(flags.JSONConfigPath)
		if err != nil {
			return Config{}, err
		}
		mergeConfigs(&flags, &jsonConfig)
	}

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return Config{}, err
	}
	mergeConfigs(&config, &flags)
	setDefaultValues(&config)

	return config, nil
}

func NewServer() (Config, error) {
	flags := parseServerFlags()

	if flags.JSONConfigPath != "" {
		jsonConfig, err := parseJSONConfig(flags.JSONConfigPath)
		if err != nil {
			return Config{}, err
		}
		mergeConfigs(&flags, &jsonConfig)
	}

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return Config{}, err
	}
	mergeConfigs(&config, &flags)

	if config.DatabaseDNS != "" {
		config.FileStoragePath = ""
		*config.StoreInterval = -1
		*config.Restore = false
	}

	return config, nil
}

func parseAgentFlags() Config {
	serverAddress := flag.String("a", "", "HTTP server endpoint address")
	reportInterval := flag.Int("r", 0, "report interval to the server (in seconds)")
	pollInterval := flag.Int("p", 0, "interval to gather metrics (in seconds)")
	key := flag.String("k", "", "key")
	rateLimit := flag.Int("l", 0, "rate limit")
	cryptoKey := flag.String("crypto-key", "", "Path to the public key")
	cfg := flag.String("config", "", "Path to JSON config file")
	flag.Parse()

	return Config{
		ServerAddress:  *serverAddress,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
		Key:            *key,
		RateLimit:      *rateLimit,
		CryptoKey:      *cryptoKey,
		JSONConfigPath: *cfg,
	}
}

func parseServerFlags() Config {
	serverAddress := flag.String("a", "", "address and port to run server")
	fileStoragePath := flag.String("f", "", "file storage path")
	databaseDSN := flag.String("d", "", "database DSN")
	restore := flag.Bool("r", false, "restore")
	storeInterval := flag.Int("i", 0, "interval")
	key := flag.String("k", "", "")
	cryptoKey := flag.String("crypto-key", "", "Path to the private key")
	cfg := flag.String("config", "", "Path to JSON config file")
	flag.Parse()

	return Config{
		ServerAddress:   *serverAddress,
		FileStoragePath: *fileStoragePath,
		DatabaseDNS:     *databaseDSN,
		Restore:         restore,
		StoreInterval:   storeInterval,
		Key:             *key,
		CryptoKey:       *cryptoKey,
		JSONConfigPath:  *cfg,
	}
}

func parseJSONConfig(filepath string) (Config, error) {
	f, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(f, &c); err != nil {
		return Config{}, err
	}

	return c, nil
}

func mergeConfigs(target, source *Config) {
	if target.ServerAddress == "" && source.ServerAddress != "" {
		target.ServerAddress = source.ServerAddress
	}
	if target.ReportInterval == 0 && source.ReportInterval != 0 {
		target.ReportInterval = source.ReportInterval
	}
	if target.PollInterval == 0 && source.PollInterval != 0 {
		target.PollInterval = source.PollInterval
	}
	if target.FileStoragePath == "" && source.FileStoragePath != "" {
		target.FileStoragePath = source.FileStoragePath
	}
	if target.Restore == nil && source.Restore != nil {
		target.Restore = source.Restore
	}
	if target.StoreInterval == nil && source.StoreInterval != nil {
		target.StoreInterval = source.StoreInterval
	}
	if target.DatabaseDNS == "" && source.DatabaseDNS != "" {
		target.DatabaseDNS = source.DatabaseDNS
	}
	if target.Key == "" && source.Key != "" {
		target.Key = source.Key
	}
	if target.RateLimit == 0 && source.RateLimit != 0 {
		target.RateLimit = source.RateLimit
	}
	if target.CryptoKey == "" && source.CryptoKey != "" {
		target.CryptoKey = source.CryptoKey
	}
}

func setDefaultValues(config *Config) {
	if config.ServerAddress == "" {
		config.ServerAddress = "localhost:8080"
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = 10
	}
	if config.PollInterval == 0 {
		config.PollInterval = 2
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = "/tmp/metrics-db.json"
	}
	if config.Restore == nil {
		restore := true
		config.Restore = &restore
	}
	if config.StoreInterval == nil {
		storeInterval := 300
		config.StoreInterval = &storeInterval
	}
	if config.RateLimit == 0 {
		config.RateLimit = 1
	}
}
