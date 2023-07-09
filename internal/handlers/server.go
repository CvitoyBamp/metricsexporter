package handlers

import (
	"flag"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"net/http"
)

type CustomServer struct {
	Endpoint string
	Server   *http.Server
	Storage  storage.IMemStorage
}

func CreateServer() *CustomServer {
	return &CustomServer{
		Server:  &http.Server{},
		Storage: storage.CreateMemStorage(),
	}
}

func (s *CustomServer) RunServer() error {
	flag.StringVar(&s.Endpoint, "a", "localhost:8080", "Add address in format host:port")
	flag.Parse()
	return http.ListenAndServe(s.Endpoint, s.MetricRouter())
}

func (s *CustomServer) StopServer() error {
	return s.Server.Close()
}
