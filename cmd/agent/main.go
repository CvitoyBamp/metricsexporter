package main

import (
	"flag"
	"github.com/CvitoyBamp/metricsexporter/internal/agent"
	"github.com/caarlos0/env/v6"
	"log"
)

func main() {

	var cfg agent.Config

	flag.IntVar(&cfg.PollInterval, "p", 2,
		"An interval for collecting metrics")
	flag.IntVar(&cfg.ReportInterval, "r", 10,
		"An interval for sending metrics to server")
	flag.StringVar(&cfg.Address, "a", "localhost:8080",
		"An address the server will send metrics")
	flag.StringVar(&cfg.Key, "k", "someKey",
		"Key for encrypting")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()

	c := agent.CreateAgent(cfg)
	c.RunAgent(cfg.PollInterval, cfg.ReportInterval)
}
