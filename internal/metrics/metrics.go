package metrics

import (
	"github.com/shirou/gopsutil/v3/mem"
	"log"
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

type LiveMetrics struct {
	sync.RWMutex
	m map[string]float64
}

func getRuntimeMetrics(m *runtime.MemStats) *LiveMetrics {
	metrics := reflect.ValueOf(m)
	liveMetrics := &LiveMetrics{
		m: make(map[string]float64),
	}

	if metrics.Kind() == reflect.Ptr {
		metrics = metrics.Elem()
	}

	liveMetrics.Lock()
	defer liveMetrics.Unlock()
	for i := 0; i < metrics.NumField(); i++ {
		metricName := metrics.Type().Field(i).Name
		metricValue := reflect.Indirect(metrics).FieldByName(metricName)
		switch metricValue.Type().Name() {
		case "float64":
			liveMetrics.m[metricName] = metricValue.Interface().(float64)
		case "uint64":
			liveMetrics.m[metricName] = float64(metricValue.Interface().(uint64))
		case "uint32":
			liveMetrics.m[metricName] = float64(metricValue.Interface().(uint32))
		default:
			continue
		}
	}

	return liveMetrics
}

func getGopsMetrics() *LiveMetrics {

	gopsMetrics := &LiveMetrics{
		m: make(map[string]float64),
	}

	v, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("can't get gops metrics, err: %v", err.Error())
	}

	//liveMetrics.Lock()
	//defer liveMetrics.Unlock()
	gopsMetrics.m["TotalMemory"] = float64(v.Total)
	gopsMetrics.m["FreeMemory"] = float64(v.Free)
	gopsMetrics.m["CPUutilization1 "] = v.UsedPercent

	return gopsMetrics
}

func (ms *Metrics) MetricGenerator(rm runtime.MemStats) {
	runtime.ReadMemStats(&rm)

	ms.Lock()
	defer ms.Unlock()
	for k, v := range getRuntimeMetrics(&rm).m {
		ms.Gauge[k] = v
	}
	for k, v := range getGopsMetrics().m {
		ms.Gauge[k] = v
	}
	ms.Gauge["RandomValue"] = rand.Float64()
	ms.Counter["PollCount"] += 1

}
