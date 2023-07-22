package handlers

import (
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/util"
	"io"
	"log"
	"net/http"
)

func (s *CustomServer) CheckAndSetMetric(metricType, metricName, metricValue string, res http.ResponseWriter) {
	// Проверка типа метрики (gauge или counter)
	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Incorrect metric type, gauge or counter is expected.", http.StatusBadRequest)
		log.Printf("Incorrect metric type recieved: %s", metricType)
	}

	err := s.Storage.SetMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(res, fmt.Sprintf("Can't parse value to %s type.", metricType), http.StatusBadRequest)
		log.Printf("Can't parse value %s to %s type.", metricValue, metricType)
	}
	res.WriteHeader(http.StatusOK)
	res.Header().Set("Content-Type", "application/json")
	log.Printf("Metric %s of type %s with value %s was successfully added", metricName, metricType, metricValue)
}

func (s *CustomServer) GetMetric(metricType, metricName string, res http.ResponseWriter, req *http.Request) {
	metricValue, err := s.Storage.GetMetric(metricType, metricName)

	if err != nil {
		log.Printf("No such metric in storage: %s", metricName)
		http.Error(res, "No such metric in storage", http.StatusNotFound)
	}

	if req.Header.Get("Content-Type") == "application/json" {
		resp, respErr := util.JsonCreator(metricValue, metricType, metricName)
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
