package config

import "flag"

type Config struct {
	Port int
}

func New() *Config {
	p := flag.Int("port", 8080, "application port")
	flag.Parse()

	return &Config{Port: *p}
}
