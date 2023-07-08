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
		return fmt.Sprintf("Value of metric %s is %d", metricName, ms.counter[metricName]), nil
	} else if metricType == "gauge" {
		return fmt.Sprintf("Value of metric %s is %f", metricName, ms.gauge[metricName]), nil
	} else {
		return "", fmt.Errorf("Don't have such metric %s of type %s.", metricName, metricType)
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
