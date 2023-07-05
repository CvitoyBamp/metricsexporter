package main

import "github.com/CvitoyBamp/metricsexporter/internal/handlers"

type SS struct {
	server *handlers.Server
}

func main() {
	server := handlers.CreateServer(":8080")
	server.HandlerRegister("/update/", server.MetricCreatorHandler)
	err := server.RunServer()
	if err != nil {
		panic("Can't start server")
	}
}
