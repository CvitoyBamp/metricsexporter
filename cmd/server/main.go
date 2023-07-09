package main

import (
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"log"
)

func main() {
	server := handlers.CreateServer()
	log.Fatal(server.RunServer())
}
