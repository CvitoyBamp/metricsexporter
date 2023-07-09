package main

import (
	"flag"
	"github.com/CvitoyBamp/metricsexporter/internal/agent"
	"log"
	"strconv"
)

var (
	addressS,
	reportIntervalS,
	pollIntervalS *string
)

func init() {
	addressS = flag.String("a", "localhost:8080", "An address the server will send metrics")
	reportIntervalS = flag.String("r", "10", "An interval for sending metrics to server")
	pollIntervalS = flag.String("p", "2", "An interval for collecting metrics")
	flag.Parse()
}

func main() {

	pollInterval, err := strconv.Atoi(*pollIntervalS)
	if err != nil {
		log.Fatalf("can't parse pollInterval. Err: %s", err)
	}

	reportInterval, err := strconv.Atoi(*reportIntervalS)
	if err != nil {
		log.Fatalf("can't parse reportInterval. Err: %s", err)
	}

	c := agent.CreateAgent(*addressS)
	c.RunAgent(pollInterval, reportInterval)
}
