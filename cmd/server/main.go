package main

import (
	"flag"
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

func main() {
	var cfg Config

	flag.StringVar(&cfg.Address, "a", "localhost:8080",
		"An address the server run")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()

	server := handlers.CreateServer()
	log.Fatal(server.RunServer(cfg.Address))
}
