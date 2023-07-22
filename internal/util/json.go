package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type JSONMetrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func JSONParser(res http.ResponseWriter, req *http.Request) *JSONMetrics {

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

func JSONCreator(metricValue, metricType, metricName string) ([]byte, error) {
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
		return json.Marshal(gData)
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
		return json.Marshal(cData)
	}
	return nil, fmt.Errorf("can't parse metric type")
}
