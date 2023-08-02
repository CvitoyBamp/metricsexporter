package main

import (
	"flag"
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"github.com/caarlos0/env/v6"
	"log"
)

const (
	user = "postgres"
	pass = "OG94Km0XV6xFJrnP"
	url  = "db.hqtykvrrpeewsakiexbk.supabase.co"
	port = "5432"
	db   = "postgres"
)

func main() {
	var cfg handlers.Config

	flag.StringVar(&cfg.Address, "a", "localhost:8080",
		"An address the server run")
	flag.IntVar(&cfg.StoreInterval, "i", 0,
		"An interval for saving metrics to file")
	flag.StringVar(&cfg.FilePath, "f", "metrics-db.json",
		"A path to save file with metrics")
	flag.BoolVar(&cfg.Restore, "r", true,
		"Boolean flag to load file with metrics")
	flag.StringVar(&cfg.DSN, "d", fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, url, port, db),
		"Database DSN")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := handlers.CreateServer(cfg)

	if cfg.Restore && cfg.FilePath != "" {
		errLoad := server.PreloadMetrics()
		if errLoad != nil {
			if fmt.Sprintf("%s", errLoad) == "EOF" {
				log.Printf("Can't read metrics from the file because it doesn't exist.")
			} else {
				log.Printf("Сan't load metrics from file, %s", errLoad)
			}
		}
	}

	if cfg.StoreInterval > 0 && cfg.FilePath != "" {
		go func() {
			server.PostSaveMetrics()
		}()
	}

	log.Fatal(server.RunServer())

}
