package main

import (
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"log"
)

func main() {
	server := handlers.CreateServer(":8080")
	log.Fatal(server.RunServer())
}
