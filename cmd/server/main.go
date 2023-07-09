package main

import (
	"flag"
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"log"
)

var (
	address *string
)

func init() {
	address = flag.String("a", "localhost:8080", "An address the server run")
}

func main() {
	flag.Parse()
	server := handlers.CreateServer()
	log.Fatal(server.RunServer(*address))
}
