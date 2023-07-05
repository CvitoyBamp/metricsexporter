package handlers

import (
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"net/http"
)

type Server struct {
	Address string
	Mux     *http.ServeMux
	Storage storage.IMemStorage
}

func CreateServer(address string) *Server {
	return &Server{
		Address: address,
		Storage: storage.CreateMemStorage(),
	}
}

func (s *Server) HandlerRegister(path string, handler http.HandlerFunc) {
	s.Mux = http.NewServeMux()
	s.Mux.HandleFunc(path, handler)
}

func (s *Server) RunServer() error {
	return http.ListenAndServe(s.Address, s.Mux)
}
