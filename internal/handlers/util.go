package handlers

import (
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/util"
	"io"
	"log"
	"net/http"
)

func (s *CustomServer) CheckAndSetMetric(metricType, metricName, metricValue string) error {
	// Проверка типа метрики (gauge или counter)
	if metricType != "gauge" && metricType != "counter" {
		log.Printf("Incorrect metric type recieved: %s", metricType)
		return fmt.Errorf("incorrect metric type, gauge or counter is expected")
	}
	return s.Storage.SetMetric(metricType, metricName, metricValue)
}

func (s *CustomServer) GetMetric(metricType, metricName string, res http.ResponseWriter, req *http.Request) {
	metricValue, err := s.Storage.GetMetric(metricType, metricName)

	if err != nil {
		log.Printf("No such metric in storage: %s", metricName)
		http.Error(res, "No such metric in storage", http.StatusNotFound)
	}

	if req.Header.Get("Content-Type") == "application/json" {
		resp, respErr := util.JSONCreator(metricValue, metricType, metricName)
		log.Println(string(resp))
		if respErr != nil {
			http.Error(res, "can't parse data as json", http.StatusBadRequest)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resp)
	} else {
		_, err = io.WriteString(res, metricValue)
		res.WriteHeader(http.StatusOK)
		if err != nil {
			log.Println("can't write answer to response")
		}
	}
}
