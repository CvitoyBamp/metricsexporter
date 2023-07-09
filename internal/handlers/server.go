package handlers

import (
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"net/http"
)

type CustomServer struct {
	Server  *http.Server
	Storage storage.IMemStorage
}

func CreateServer() *CustomServer {
	return &CustomServer{
		Server:  &http.Server{},
		Storage: storage.CreateMemStorage(),
	}
}

func (s *CustomServer) RunServer(address string) error {
	return http.ListenAndServe(address, s.MetricRouter())
}

func (s *CustomServer) StopServer() error {
	return s.Server.Close()
}
