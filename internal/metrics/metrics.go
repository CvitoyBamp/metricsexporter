package metrics

import (
	"math/rand"
	"reflect"
	"runtime"
	"sync"
)

type Metrics struct {
	sync.Mutex
	Gauge   map[string]float64
	Counter map[string]int64
}

func getRuntimeMetrics(m *runtime.MemStats) map[string]float64 {
	metrics := reflect.ValueOf(m)
	runtimeMetrics := make(map[string]float64)

	if metrics.Kind() == reflect.Ptr {
		metrics = metrics.Elem()
	}

	for i := 0; i < metrics.NumField(); i++ {
		metricName := metrics.Type().Field(i).Name
		metricValue := reflect.Indirect(metrics).FieldByName(metricName)
		switch metricValue.Type().Name() {
		case "float64":
			runtimeMetrics[metricName] = metricValue.Interface().(float64)
		case "uint64":
			runtimeMetrics[metricName] = float64(metricValue.Interface().(uint64))
		case "uint32":
			runtimeMetrics[metricName] = float64(metricValue.Interface().(uint32))
		default:
			continue
		}
	}

	return runtimeMetrics
}

func (ms *Metrics) MetricGenerator(rm runtime.MemStats) *Metrics {
	runtime.ReadMemStats(&rm)

	for k, v := range getRuntimeMetrics(&rm) {
		ms.Gauge[k] = v
	}
	ms.Gauge["RandomValue"] = rand.Float64()
	ms.Counter["PollCount"] += 1

	return ms
}
