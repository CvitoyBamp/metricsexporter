package handlers

import (
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"net/http"
)

type CustomServer struct {
	Server  *http.Server
	Address string
	Mux     *http.ServeMux
	Storage storage.IMemStorage
}

func CreateServer(address string) *CustomServer {
	return &CustomServer{
		Address: address,
		Storage: storage.CreateMemStorage(),
	}
}

func (s *CustomServer) HandlerRegister(path string, handler http.HandlerFunc) {
	s.Mux = http.NewServeMux()
	s.Mux.HandleFunc(path, handler)
}

func (s *CustomServer) RunServer() error {
	return http.ListenAndServe(s.Address, s.Mux)
}

func (s *CustomServer) StopServer() error {
	return s.Server.Close()
}
