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
	SetMetric(name string, m *Metric) MemStorage
	GetMetric(name string) Metric
	DeleteMetric(name string) MemStorage
}

func (ms *MemStorage) SetMetric(name string, m *Metric) MemStorage {
	ms.metrics[name] = *m
	return *ms
}

func (ms *MemStorage) GetMetric(name string) Metric {
	return ms.metrics[name]
}

func (ms *MemStorage) DeleteMetric(name string) MemStorage {
	delete(ms.metrics, name)
	return *ms
}

func metricCreatorHandler(res http.ResponseWriter, req *http.Request) {

	var metric Metric

	ms := &MemStorage{metrics: make(map[string]Metric)}

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

	// Проверка типа метрики (gauge или counter)
	if path[1] != "gauge" && path[1] != "counter" {
		http.Error(res, "Incorrect metric type, gauge or counter is expected.", http.StatusBadRequest)
		return
	} else if path[1] == "counter" {
		value, err := strconv.ParseInt(path[3], 10, 64)
		metric = Metric{MName: path[2], MType: "counter", Vcounter: value}
		if err != nil {
			http.Error(res, "Can't parse value to counter type (int64)", http.StatusBadRequest)
		}
		ms.SetMetric(path[2], &metric)
		return
	} else if path[1] == "gauge" {
		value, err := strconv.ParseFloat(path[3], 64)
		metric = Metric{MName: path[2], MType: "counter", Vgauge: value}
		if err != nil {
			http.Error(res, "Can't parse value to gauge type (float64)", http.StatusBadRequest)
		}
		ms.SetMetric(path[2], &metric)
		return
	}

	res.WriteHeader(http.StatusOK)
	return
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", metricCreatorHandler)

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic("Can't create server :(")
	}
}
