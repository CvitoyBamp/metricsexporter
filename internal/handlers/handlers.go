package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	cors2 "github.com/go-chi/cors"
	"html/template"
	"io"
	"log"
	"net/http"
)

type MetricsList struct {
	MetricName  string
	MetricValue string
}

var htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Metrics list</title>
</head>
<body>
{{ range . }}
    <p>Metric name: {{ .MetricName }}</p>
    <p>Metric value: {{ .MetricValue }}</p>
    <span>----------------</span>
{{ end }}
</body>
</html>
`

func (s *CustomServer) MetricRouter() chi.Router {
	r := chi.NewRouter()
	cors := cors2.New(cors2.Options{
		AllowedMethods: []string{http.MethodPost, http.MethodGet},
		AllowedHeaders: []string{"Content-Type"},
	})

	r.Group(func(r chi.Router) {
		r.Use(cors.Handler)
		r.Use(middleware.Logger)
		r.Route("/", func(r chi.Router) {
			r.Get("/", s.GetAllMetricsHandler)
			r.Route("/value", func(r chi.Router) {
				r.Get("/{metricType}/{metricName}", s.GetMetricValueHandler)
			})
			r.Route("/update", func(r chi.Router) {
				r.Post("/{metricType}/{metricName}/{metricValue}", s.MetricCreatorHandler)
			})
		})
	})

	return r
}

func (s *CustomServer) GetAllMetricsHandler(res http.ResponseWriter, req *http.Request) {

	var metricList []MetricsList
	var metric MetricsList
	list, err := s.Storage.GetExistsMetrics()

	if err != nil {
		http.Error(res, "No metrics in storage", http.StatusNotFound)
		return
	}

	for k, v := range list {
		metric.MetricValue = v
		metric.MetricName = k
		metricList = append(metricList, metric)
	}

	fmt.Println(metricList)

	tpl := template.New("Metrics Page")
	tmpl, err := tpl.Parse(htmlTemplate)

	tmpl.Execute(res, metricList)
}

func (s *CustomServer) GetMetricValueHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "metricType")
	metricName := chi.URLParam(req, "metricName")

	metricValue, err := s.Storage.GetMetric(metricType, metricName)

	if err != nil {
		log.Printf("No such metric in storage: %s", metricName)
		http.Error(res, "No such metric in storage", http.StatusNotFound)
		return
	}

	res.WriteHeader(http.StatusOK)
	io.WriteString(res, metricValue)
	return

}

func (s *CustomServer) MetricCreatorHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "metricType")
	metricName := chi.URLParam(req, "metricName")
	metricValue := chi.URLParam(req, "metricValue")

	// Проверка типа метрики (gauge или counter)
	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Incorrect metric type, gauge or counter is expected.", http.StatusBadRequest)
		log.Printf("Incorrect metric type recieved: %s", metricType)
		return
	} else if metricType == "counter" {
		err := s.Storage.SetMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, "Can't parse value to counter type (int64)", http.StatusBadRequest)
			log.Printf("Can't parse value %s to counter type (int64)", metricValue)
			return
		}
		log.Printf("Metric %s of type %s with value %s was successfully added", metricName, metricType, metricValue)
		return
	} else if metricType == "gauge" {
		err := s.Storage.SetMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, "Can't parse value to gauge type (float64)", http.StatusBadRequest)
			log.Printf("Can't parse value %s to gauge type (float64)", metricValue)
			return
		}
		log.Printf("Metric %s of type %s with value %s was successfully added", metricName, metricType, metricValue)
		return
	}
	res.WriteHeader(http.StatusOK)
	return
}
