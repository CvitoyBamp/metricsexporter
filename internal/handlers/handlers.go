package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) MetricCreatorHandler(res http.ResponseWriter, req *http.Request) {

	// Массив с path-параметрами из URL'а
	path := strings.Split(req.URL.Path[1:], "/")

	// Проверка метода запроса
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST-requests available.", http.StatusMethodNotAllowed)
		return
	}

	// Проверка корректности URL'а
	if len(path) != 4 {
		http.Error(res, "Expected format: "+
			"\rhttp://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", http.StatusNotFound)
		return
	}

	metricType, metricName, metricValue := path[1], path[2], path[3]

	// Проверка типа метрики (gauge или counter)
	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Incorrect metric type, gauge or counter is expected.", http.StatusBadRequest)
		return
	} else if metricType == "counter" {
		err := s.Storage.SetMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, "Can't parse value to counter type (int64)", http.StatusBadRequest)
		}
		fmt.Fprintf(res, "Counter metric %s with values %s was added.", path[2], path[3])
		fmt.Printf("Counter metric %s with values %s was added.", path[2], path[3])
		return
	} else if metricType == "gauge" {
		err := s.Storage.SetMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, "Can't parse value to gauge type (float64)", http.StatusBadRequest)
		}
		fmt.Fprintf(res, "Gauge metric %s with values %s was added.", path[2], path[3])
		fmt.Printf("Counter metric %s with values %s was added.", path[2], path[3])
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
	return
}
