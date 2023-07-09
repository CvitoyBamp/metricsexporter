package storage

import (
	"fmt"
	"strconv"
)

// MemStorage In-memory хранилище метрик
type MemStorage struct {
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
			return fmt.Errorf("Can't parse value to counter type (int64), error: %s", err)
		}
		ms.counter[metricName] = value
	} else if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("Can't parse value to gauge type (float64), error: %s", err)
		}
		ms.gauge[metricName] = value
	} else {
		return fmt.Errorf("Don't know such type: %s", metricType)
	}

	return nil
}

func (ms *MemStorage) GetMetric(metricType, metricName string) (string, error) {
	if metricType == "counter" {
		_, ok := ms.counter[metricName]
		if ok {
			return fmt.Sprintf("%d", ms.counter[metricName]), nil
		} else {
			return "", fmt.Errorf("Don't have metric %s of type %s in storage.", metricName, metricType)
		}
	} else if metricType == "gauge" {
		_, ok := ms.gauge[metricName]
		if ok {
			return fmt.Sprintf("%f", ms.gauge[metricName]), nil
		} else {
			return "", fmt.Errorf("Don't have metric %s of type %s in storage.", metricName, metricType)
		}
	} else {
		return "", fmt.Errorf("Don't have metric's type %s in storage.", metricType)
	}
}

func (ms *MemStorage) GetExistsMetrics() (map[string]string, error) {
	l := len(ms.gauge) + len(ms.counter)
	if l != 0 {
		metricsList := make(map[string]string, l)
		for k, v := range ms.gauge {
			metricsList[k] = fmt.Sprintf("%f", v)
		}
		for k, v := range ms.counter {
			metricsList[k] = fmt.Sprintf("%d", v)
		}
		return metricsList, nil
	} else {
		return nil, fmt.Errorf("No metrics in storage for now.")
	}
}

func (ms *MemStorage) DeleteMetric(metricType, metricName string) error {
	if metricType == "counter" {
		delete(ms.counter, metricName)
	} else if metricType == "gauge" {
		delete(ms.gauge, metricName)
	} else {
		fmt.Errorf("Don't have such metric %s of type %s.", metricName, metricType)
	}
	return nil
}
