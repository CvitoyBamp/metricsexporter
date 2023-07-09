package handlers

import (
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"net/http"
)

type CustomServer struct {
	Server  *http.Server
	Storage storage.IMemStorage
}

func CreateServer(address string) *CustomServer {
	return &CustomServer{
		Server: &http.Server{
			Addr: address,
		},
		Storage: storage.CreateMemStorage(),
	}
}

func (s *CustomServer) RunServer() error {
	return http.ListenAndServe(s.Server.Addr, s.MetricRouter())
}

func (s *CustomServer) StopServer() error {
	return s.Server.Close()
}
