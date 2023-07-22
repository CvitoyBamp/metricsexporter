package handlers

import (
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/util"
	"github.com/go-chi/chi/v5"
	cors2 "github.com/go-chi/cors"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type MetricsList struct {
	MetricName  string
	MetricValue string
}

type JSONMetrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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
		r.Route("/", func(r chi.Router) {
			r.Get("/", Logging(s.GetAllMetricsHandler()))
			r.Route("/value", func(r chi.Router) {
				r.Post("/", Logging(s.GetJSONMetricHandler()))
				r.Get("/{metricType}/{metricName}", Logging(s.GetMetricValueHandler()))
			})
			r.Route("/update", func(r chi.Router) {
				r.Post("/", Logging(s.CreateJSONMetricHandler()))
				r.Post("/{metricType}/{metricName}/{metricValue}", Logging(s.MetricCreatorHandler()))
			})
		})
	})

	return r
}

func (s *CustomServer) GetAllMetricsHandler() http.Handler {
	fn := func(res http.ResponseWriter, _ *http.Request) {
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
		if err != nil {
			log.Print("can't parse template")
			http.Error(res, "can't parse template", http.StatusInternalServerError)
			return
		}

		tmplerr := tmpl.Execute(res, metricList)
		if tmplerr != nil {
			log.Print("can't create template")
			http.Error(res, "can't parse template", http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) GetMetricValueHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		s.GetMetric(metricType, metricName, res, req)
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) MetricCreatorHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		metricValue := chi.URLParam(req, "metricValue")
		err := s.CheckAndSetMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, fmt.Sprintf("%s.", err), http.StatusBadRequest)
			return
		}
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) CreateJSONMetricHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {

		data := util.JSONParser(res, req)
		log.Println(*data.Value)

		if data.MType != "gauge" && data.MType != "counter" {
			http.Error(res, "Incorrect metric type, gauge or counter is expected.", http.StatusBadRequest)
			log.Println("Incorrect metric type, gauge or counter is expected.")
			return
		} else if data.MType == "gauge" {
			err := s.CheckAndSetMetric(data.MType, data.ID, strconv.FormatFloat(*data.Value, 'f', -1, 64))
			if err != nil {
				http.Error(res, fmt.Sprintf("%s.", err), http.StatusBadRequest)
				log.Println("can't add metric to storage")
				return
			}
		} else if data.MType == "counter" {
			err := s.CheckAndSetMetric(data.MType, data.ID, strconv.FormatInt(*data.Delta, 10))
			if err != nil {
				http.Error(res, fmt.Sprintf("%s.", err), http.StatusBadRequest)
				log.Println("can't add metric to storage")
				return
			}
		}

		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json")
		log.Printf("Metric %s of type %s was successfully added", data.ID, data.MType)
		return

	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) GetJSONMetricHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		data := util.JSONParser(res, req)
		s.GetMetric(data.MType, data.ID, res, req)
	}
	return http.HandlerFunc(fn)
}
