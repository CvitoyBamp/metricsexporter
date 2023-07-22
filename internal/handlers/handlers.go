package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	cors2 "github.com/go-chi/cors"
	"html/template"
	"io"
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
		}

		tmplerr := tmpl.Execute(res, metricList)
		if tmplerr != nil {
			log.Print("can't create template")
			http.Error(res, "can't parse template", http.StatusInternalServerError)
		}

		res.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) GetMetricValueHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		s.getMetric(metricType, metricName, res, req)
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) MetricCreatorHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		metricValue := chi.URLParam(req, "metricValue")
		s.checkAndSetMetric(metricType, metricName, metricValue, res)
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) CreateJSONMetricHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {

		data := s.jsonParser(res, req)

		if data.MType != "gauge" && data.MType != "counter" {
			http.Error(res, "Incorrect metric type, gauge or counter is expected.", http.StatusBadRequest)
			log.Printf("Incorrect metric type recieved: %s", data.MType)
			return
		} else if data.MType == "gauge" {
			s.checkAndSetMetric(data.MType, data.ID, fmt.Sprintf("%f", *data.Value), res)
		} else if data.MType == "counter" {
			s.checkAndSetMetric(data.MType, data.ID, fmt.Sprintf("%d", *data.Delta), res)
		}
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) GetJSONMetricHandler() http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		data := s.jsonParser(res, req)
		s.getMetric(data.MType, data.ID, res, req)
	}
	return http.HandlerFunc(fn)
}

func (s *CustomServer) checkAndSetMetric(metricType, metricName, metricValue string, res http.ResponseWriter) {
	// Проверка типа метрики (gauge или counter)
	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Incorrect metric type, gauge or counter is expected.", http.StatusBadRequest)
		log.Printf("Incorrect metric type recieved: %s", metricType)
		return
	}

	err := s.Storage.SetMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(res, fmt.Sprintf("Can't parse value to %s type.", metricType), http.StatusBadRequest)
		log.Printf("Can't parse value %s to %s type.", metricValue, metricType)
		return
	}
	res.WriteHeader(http.StatusOK)
	log.Printf("Metric %s of type %s with value %s was successfully added", metricName, metricType, metricValue)
	return
}

func (s *CustomServer) getMetric(metricType, metricName string, res http.ResponseWriter, req *http.Request) {
	metricValue, err := s.Storage.GetMetric(metricType, metricName)

	if err != nil {
		log.Printf("No such metric in storage: %s", metricName)
		http.Error(res, "No such metric in storage", http.StatusNotFound)
		return
	}

	res.WriteHeader(http.StatusOK)

	if req.Header.Get("Content-Type") == "application/json" {
		s.jsonCreator(metricValue, metricType, metricName, res)
	} else {
		_, err = io.WriteString(res, metricValue)
		if err != nil {
			log.Println("can't write answer to response")
		}
	}
}

func (s *CustomServer) jsonParser(res http.ResponseWriter, req *http.Request) *JSONMetrics {

	var jsonStruct JSONMetrics

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Only application/json supported.", http.StatusUnsupportedMediaType)
	}

	body, errReq := io.ReadAll(req.Body)
	if errReq != nil {
		http.Error(res, "Can't read request body", http.StatusBadRequest)
	}

	err := json.Unmarshal(body, &jsonStruct)
	if err != nil {
		http.Error(res, "can't parse json", http.StatusBadRequest)
	}

	return &jsonStruct
}

func (s *CustomServer) jsonCreator(metricValue, metricType, metricName string, res http.ResponseWriter) {
	var cData JSONMetrics
	var gData JSONMetrics

	if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			_ = fmt.Errorf("can't parse value to gauge type (float64), error: %s", err)
		}

		gData = JSONMetrics{
			ID:    metricName,
			MType: metricType,
			Value: &value,
		}
		resp, respErr := json.Marshal(gData)
		if respErr != nil {
			http.Error(res, "can't parse data as json", http.StatusBadRequest)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resp)
	}

	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			_ = fmt.Errorf("can't parse value to counter type (int64), error: %s", err)
		}
		cData = JSONMetrics{
			ID:    metricName,
			MType: metricType,
			Delta: &value,
		}
		resp, respErr := json.Marshal(cData)
		if respErr != nil {
			http.Error(res, "can't parse data as json", http.StatusBadRequest)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resp)
	}

}
