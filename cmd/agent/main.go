package main

import (
	"flag"
	"github.com/CvitoyBamp/metricsexporter/internal/agent"
)

func main() {
	address := flag.String("a", "localhost:8080", "An address the server will send metrics")
	reportInterval := flag.Int("r", 10, "An interval for sending metrics to server")
	pollInterval := flag.Int("p", 2, "An interval for collecting metrics")
	flag.Parse()
	c := agent.CreateAgent(*address)
	c.RunAgent(*pollInterval, *reportInterval)
}
