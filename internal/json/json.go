package json

import (
	"encoding/json"
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func Parser(res http.ResponseWriter, req *http.Request) *Metrics {

	var jsonStruct Metrics

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

func Creator(metricValue, metricType, metricName string) ([]byte, error) {
	var cData Metrics
	var gData Metrics

	if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			_ = fmt.Errorf("can't parse value to gauge type (float64), error: %s", err)
		}

		gData = Metrics{
			ID:    metricName,
			MType: metricType,
			Value: &value,
		}
		return json.Marshal(gData)
	}

	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			_ = fmt.Errorf("can't parse value to counter type (int64), error: %s", err)
		}
		cData = Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &value,
		}
		return json.Marshal(cData)
	}
	return nil, fmt.Errorf("can't parse metric type")
}

func MetricConverter(ms *storage.MemStorage) ([]byte, error) {

	//ms.RLock()
	//defer ms.RUnlock()

	var arr []string
	var metric Metrics

	for k, v := range ms.GetGaugeMetrics() {
		metric = Metrics{
			ID:    k,
			MType: "gauge",
			Value: &v,
		}
		data, err := json.Marshal(metric)
		if err != nil {
			return nil, err
		}
		arr = append(arr, string(data))
	}

	for k, v := range ms.GetCounterMetrics() {
		metric = Metrics{
			ID:    k,
			MType: "counter",
			Delta: &v,
		}
		data, err := json.Marshal(metric)
		if err != nil {
			return nil, err
		}
		arr = append(arr, string(data))
	}

	output := "[" + strings.Join(arr, ",") + "]"

	return []byte(output), nil
}

func Decoder(data []byte, ms *storage.MemStorage) error {
	var jsonData []Metrics
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return err
	}

	for _, v := range jsonData {
		if v.MType == "gauge" {
			errSetG := ms.SetMetric(v.MType, v.ID, strconv.FormatFloat(*v.Value, 'f', -1, 64))
			if errSetG != nil {
				return errSetG
			}
		}
		if v.MType == "counter" {
			errSetC := ms.SetMetric(v.MType, v.ID, strconv.FormatInt(*v.Delta, 10))
			if errSetC != nil {
				return errSetC
			}
		}
	}

	return nil
}

func ListParser(res http.ResponseWriter, req *http.Request) []Metrics {

	var jsonStruct []Metrics

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

	return jsonStruct
}

func ListCreator(gm map[string]float64, cm map[string]int64) ([]byte, error) {

	var metrics Metrics
	var metricsList []string

	for k, v := range gm {
		metrics = Metrics{
			ID:    k,
			MType: "gauge",
			Value: &v,
		}
		data, err := json.Marshal(metrics)
		if err != nil {
			return nil, err
		}
		metricsList = append(metricsList, string(data))
	}

	for k, v := range cm {
		metrics = Metrics{
			ID:    k,
			MType: "counter",
			Delta: &v,
		}
		data, err := json.Marshal(metrics)
		if err != nil {
			return nil, err
		}
		metricsList = append(metricsList, string(data))
	}

	output := "[" + strings.Join(metricsList, ",") + "]"

	return []byte(output), nil
}
