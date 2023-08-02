package handlers

import (
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/db"
	"github.com/CvitoyBamp/metricsexporter/internal/json"
	"github.com/CvitoyBamp/metricsexporter/internal/middlewares"
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
		AllowedHeaders: []string{"Content-Type", "Content-Encoding", "Accept-Encoding"},
	})

	r.Group(func(r chi.Router) {
		r.Use(cors.Handler)
		r.Use(middlewares.MiddlewareZIP)
		//r.Use(middleware.Compress(5, "application/json", "text/html; charset=UTF-8"))
		r.Route("/", func(r chi.Router) {
			r.Get("/", middlewares.Logging(s.getAllMetricsHandler()))
			r.Get("/ping", s.checkDBConnectivityHandler)
			r.Route("/value", func(r chi.Router) {
				r.Post("/", middlewares.Logging(s.getJSONMetricHandler()))
				r.Get("/{metricType}/{metricName}", middlewares.Logging(s.getMetricValueHandler()))
			})
			r.Route("/update", func(r chi.Router) {
				r.Post("/", middlewares.Logging(s.createJSONMetricHandler()))
				r.Post("/{metricType}/{metricName}/{metricValue}", middlewares.Logging(s.metricCreatorHandler()))
			})
		})
	})

	return r
}

func (s *CustomServer) getAllMetricsHandler() http.Handler {
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

		tpl := template.New("Metrics Page")
		tmpl, err := tpl.Parse(htmlTemplate)
		if err != nil {
			log.Print("can't parse template")
			http.Error(res, "can't parse template", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")

		tmplerr := tmpl.Execute(res, metricList)
		if tmplerr != nil {
			log.Print("can't create template")
			http.Error(res, "can't parse template", http.StatusInternalServerError)
			return
		}

	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) getMetricValueHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		s.GetMetric(metricType, metricName, res, req)
		res.Header().Set("Content-Type", "text/plain")
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) metricCreatorHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		metricValue := chi.URLParam(req, "metricValue")
		err := s.CheckAndSetMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, fmt.Sprintf("%s.", err), http.StatusBadRequest)
			return
		}
		if s.Config.StoreInterval == 0 && s.Config.FilePath != "" {
			s.SyncSavingToFile()
		}
		res.Header().Set("Content-Type", "text/plain")
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) createJSONMetricHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {

		data := json.Parser(res, req)

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
		if s.Config.StoreInterval == 0 && s.Config.FilePath != "" {
			s.SyncSavingToFile()
		}
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json")
		log.Printf("Metric %s of type %s was successfully added", data.ID, data.MType)
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) getJSONMetricHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		data := json.Parser(res, req)
		s.GetMetric(data.MType, data.ID, res, req)
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) checkDBConnectivityHandler(res http.ResponseWriter, req *http.Request) {
	err := db.CheckConnectivity(s.Config.DSN)
	if err != nil {
		http.Error(res, fmt.Sprintf("%s", err), http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}
