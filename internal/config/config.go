// package config

// import "flag"

// type Config struct {
// 	Port int
// }

// func New() *Config {
// 	p := flag.Int("port", 8080, "application port")
// 	flag.Parse()

// 	return &Config{Port: *p}
// }

package config

// import (
// 	"flag"
// )

// var serverAddress string
// var reportInterval int
// var pollInterval int

// func ParseAgentFlags() (*string, *int, *int) {
// 	serverAddress := flag.String("a", "localhost:8080", "HTTP server endpoint address")
// 	reportInterval:=flag.Int("r", 10, "report interval to the server (in seconds)")
// 	pollInterval:=flag.Int("p", 2, "interval to gather metrics (in seconds)")

// 	flag.Parse()
// }

// func ParseFlags() *string {
// 	serverAddress:=flag.String("a", "localhost:8080", "address and port to run server")
// 	flag.Parse()

// 	return serverAddress
// }
