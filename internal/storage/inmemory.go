package storage

import (
	"fmt"
	"strconv"
	"sync"
)

// MemStorage In-memory хранилище метрик
type MemStorage struct {
	sync.RWMutex
	Gauge   map[string]float64
	Counter map[string]int64
}

func CreateMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (ms *MemStorage) SetMetric(metricType, metricName, metricValue string) error {
	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse value to counter type (int64), error: %s", err)
		}
		ms.Lock()
		ms.Counter[metricName] += value
		ms.Unlock()
	} else if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("can't parse value to gauge type (float64), error: %s", err)
		}
		ms.Lock()
		ms.Gauge[metricName] = value
		ms.Unlock()
	} else {
		return fmt.Errorf("don't know such type: %s", metricType)
	}

	return nil
}

func (ms *MemStorage) GetMetric(metricType, metricName string) (string, error) {
	if metricType == "counter" {
		ms.RLock()
		defer ms.RUnlock()
		_, ok := ms.Counter[metricName]
		if ok {
			val := ms.Counter[metricName]
			return fmt.Sprintf("%d", val), nil
		} else {
			return "", fmt.Errorf("don't have metric %s of type %s in storage", metricName, metricType)
		}
	} else if metricType == "gauge" {
		_, ok := ms.Gauge[metricName]
		if ok {
			val := ms.Gauge[metricName]
			return strconv.FormatFloat(val, 'f', -1, 64), nil
		} else {
			return "", fmt.Errorf("don't have metric %s of type %s in storage", metricName, metricType)
		}
	} else {
		return "", fmt.Errorf("don't have metric's type %s in storage", metricType)
	}
}

func (ms *MemStorage) GetExistsMetrics() (map[string]string, error) {
	l := len(ms.Gauge) + len(ms.Counter)
	if l != 0 {
		metricsList := make(map[string]string, l)
		ms.Lock()
		defer ms.Unlock()
		for k, v := range ms.Gauge {
			metricsList[k] = fmt.Sprintf("%f", v)
		}
		for k, v := range ms.Counter {
			metricsList[k] = fmt.Sprintf("%d", v)
		}
		return metricsList, nil
	} else {
		return nil, fmt.Errorf("no metrics in storage for now")
	}
}

func (ms *MemStorage) DeleteMetric(metricType, metricName string) error {
	if metricType == "counter" {
		delete(ms.Counter, metricName)
	} else if metricType == "gauge" {
		delete(ms.Gauge, metricName)
	} else {
		_ = fmt.Errorf("don't have such metric %s of type %s", metricName, metricType)
	}
	return nil
}
