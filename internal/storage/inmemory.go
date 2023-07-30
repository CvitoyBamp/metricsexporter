package storage

import (
	"fmt"
	"strconv"
	"sync"
)

// MemStorage In-memory хранилище метрик
type MemStorage struct {
	sync.RWMutex
	gauge   map[string]float64
	counter map[string]int64
}

func CreateMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (ms *MemStorage) SetMetric(metricType, metricName, metricValue string) error {
	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse value to counter type (int64), error: %s", err)
		}
		ms.Lock()
		ms.counter[metricName] += value
		ms.Unlock()
	} else if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("can't parse value to gauge type (float64), error: %s", err)
		}
		ms.Lock()
		ms.gauge[metricName] = value
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
		_, ok := ms.counter[metricName]
		if ok {
			val := ms.counter[metricName]
			return fmt.Sprintf("%d", val), nil
		} else {
			return "", fmt.Errorf("don't have metric %s of type %s in storage", metricName, metricType)
		}
	} else if metricType == "gauge" {
		_, ok := ms.gauge[metricName]
		if ok {
			val := ms.gauge[metricName]
			return strconv.FormatFloat(val, 'f', -1, 64), nil
		} else {
			return "", fmt.Errorf("don't have metric %s of type %s in storage", metricName, metricType)
		}
	} else {
		return "", fmt.Errorf("don't have metric's type %s in storage", metricType)
	}
}

func (ms *MemStorage) GetExistsMetrics() (map[string]string, error) {
	l := len(ms.gauge) + len(ms.counter)
	if l != 0 {
		metricsList := make(map[string]string, l)
		ms.Lock()
		defer ms.Unlock()
		for k, v := range ms.gauge {
			metricsList[k] = fmt.Sprintf("%f", v)
		}
		for k, v := range ms.counter {
			metricsList[k] = fmt.Sprintf("%d", v)
		}
		return metricsList, nil
	} else {
		return nil, fmt.Errorf("no metrics in storage for now")
	}
}

func (ms *MemStorage) GetGaugeMetrics() map[string]float64 {
	return ms.gauge
}

func (ms *MemStorage) GetCounterMetrics() map[string]int64 {
	return ms.counter
}

func (ms *MemStorage) DeleteMetric(metricType, metricName string) error {
	if metricType == "counter" {
		delete(ms.counter, metricName)
	} else if metricType == "gauge" {
		delete(ms.gauge, metricName)
	} else {
		_ = fmt.Errorf("don't have such metric %s of type %s", metricName, metricType)
	}
	return nil
}
