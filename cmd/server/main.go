package main

import (
	"flag"
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
}

func main() {
	var cfg Config

	flag.StringVar(&cfg.Address, "a", "localhost:8080",
		"An address the server run")
	flag.IntVar(&cfg.StoreInterval, "i", 10,
		"An interval for saving metrics to file")
	flag.StringVar(&cfg.FilePath, "f", "metrics-db.json",
		"A path to save file with metrics")
	flag.BoolVar(&cfg.Restore, "r", true,
		"Boolean flag to load file with metrics")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()

	server := handlers.CreateServer()

	if cfg.Restore {
		errLoad := server.PreloadMetrics(cfg.FilePath)
		if errLoad != nil {
			log.Printf("can't load metrics from file, %s", errLoad)
		}
	}

	go func() {
		log.Fatal(server.PostSaveMetrics(cfg.FilePath, cfg.StoreInterval))
	}()

	log.Fatal(server.RunServer(cfg.Address))

}
