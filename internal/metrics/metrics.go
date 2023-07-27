package metrics

import (
	"math/rand"
	"reflect"
	"runtime"
	"sync"
)

type Metrics struct {
	sync.RWMutex
	Gauge   map[string]float64
	Counter map[string]int64
}

type RuntimeMetrics struct {
	sync.RWMutex
	m map[string]float64
}

func getRuntimeMetrics(m *runtime.MemStats) *RuntimeMetrics {
	metrics := reflect.ValueOf(m)
	runtimeMetrics := &RuntimeMetrics{
		m: make(map[string]float64),
	}

	if metrics.Kind() == reflect.Ptr {
		metrics = metrics.Elem()
	}

	runtimeMetrics.Lock()
	defer runtimeMetrics.Unlock()
	for i := 0; i < metrics.NumField(); i++ {
		metricName := metrics.Type().Field(i).Name
		metricValue := reflect.Indirect(metrics).FieldByName(metricName)
		switch metricValue.Type().Name() {
		case "float64":
			runtimeMetrics.m[metricName] = metricValue.Interface().(float64)
		case "uint64":
			runtimeMetrics.m[metricName] = float64(metricValue.Interface().(uint64))
		case "uint32":
			runtimeMetrics.m[metricName] = float64(metricValue.Interface().(uint32))
		default:
			continue
		}
	}

	return runtimeMetrics
}

func (ms *Metrics) MetricGenerator(rm runtime.MemStats) *Metrics {
	runtime.ReadMemStats(&rm)

	ms.Lock()
	for k, v := range getRuntimeMetrics(&rm).m {
		ms.Gauge[k] = v
	}
	ms.Gauge["RandomValue"] = rand.Float64()
	ms.Counter["PollCount"] += 1
	ms.Unlock()

	return ms
}
