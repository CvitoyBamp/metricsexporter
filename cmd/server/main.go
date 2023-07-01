package main

import (
	"net/http"
	"strconv"
	"strings"
)

// Metric Определяю абстрацию метрики
type Metric struct {
	MName    string  // имя метрики
	MType    string  // тип метрики
	Vgauge   float64 // значение метрики при типе gauge
	Vcounter int64   // значение метрики при типе counter
}

// MemStorage In-memory хранилище метрик
type MemStorage struct {
	metrics map[string]Metric
}

// IMemStorage Интрфейс с абстрактным функциями добавления, просмотра и удаления метрик в хранилище
type IMemStorage interface {
	SetMetric(name string, m Metric) MemStorage
	GetMetric(name string) Metric
	DeleteMetric(name string) MemStorage
}

func (ms MemStorage) SetMetric(name string, m Metric) MemStorage {
	ms.metrics[name] = m
	return ms
}

func (ms MemStorage) GetMetric(name string) Metric {
	return ms.metrics[name]
}

func (ms MemStorage) DeleteMetric(name string) MemStorage {
	delete(ms.metrics, name)
	return ms
}

func metricCreatorHandler(res http.ResponseWriter, req *http.Request) {

	var metric Metric

	ms := &MemStorage{metrics: make(map[string]Metric)}

	// Массив с path-параметрами из URL'а
	path := strings.Split(req.URL.Path[1:], "/")

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		res.Write([]byte("Only POST-requests available."))
		return
	}

	if req.Header.Get("Content-Type") != "text/plain" {
		res.WriteHeader(http.StatusUnsupportedMediaType)
		res.Write([]byte("Unsupported Media Type, should be text/plain."))
		return
	}

	if len(path) != 4 {
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte("Expected format: " +
			"\rhttp://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>"))
	}

	if path[1] != "gauge" && path[1] != "counter" {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Incorrect metric type, gauge or counter is expected."))
	} else if path[1] == "counter" {
		value, err := strconv.ParseInt(path[3], 10, 64)
		metric = Metric{MName: path[2], MType: "counter", Vcounter: value}
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Can't parse value to counter type (int64)"))
		}
	} else if path[1] == "gauge" {
		value, err := strconv.ParseFloat(path[3], 64)
		metric = Metric{MName: path[2], MType: "counter", Vgauge: value}
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Can't parse value to gauge type (float64)"))
		}
	}

	ms.SetMetric(path[2], metric)
	res.WriteHeader(http.StatusOK)

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", metricCreatorHandler)

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic("Can't create server :(")
	}
}
