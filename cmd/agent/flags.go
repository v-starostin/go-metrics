package main

import (
	"flag"
)

var serverAddress string
var reportInterval int
var pollInterval int

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server endpoint address")
	flag.IntVar(&reportInterval, "r", 10, "report interval to the server (in seconds)")
	flag.IntVar(&pollInterval, "p", 2, "interval to gather metrics (in seconds)")

	flag.Parse()
}
