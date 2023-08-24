package metrics

import (
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"reflect"
	"runtime"
)

func getRuntimeMetrics(m *runtime.MemStats, liveMetrics *storage.MemStorage) {
	metrics := reflect.ValueOf(m)

	if metrics.Kind() == reflect.Ptr {
		metrics = metrics.Elem()
	}

	//liveMetrics.Lock()
	//defer liveMetrics.Unlock()
	for i := 0; i < metrics.NumField(); i++ {
		metricName := metrics.Type().Field(i).Name
		metricValue := reflect.Indirect(metrics).FieldByName(metricName)
		switch metricValue.Type().Name() {
		case "float64":
			liveMetrics.Gauge[metricName] = metricValue.Interface().(float64)
		case "uint64":
			liveMetrics.Gauge[metricName] = float64(metricValue.Interface().(uint64))
		case "uint32":
			liveMetrics.Gauge[metricName] = float64(metricValue.Interface().(uint32))
		default:
			continue
		}
	}
}

func getGopsMetrics(liveMetrics *storage.MemStorage) {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("can't get gops metrics, err: %v", err.Error())
	}

	//liveMetrics.Lock()
	//defer liveMetrics.Unlock()
	liveMetrics.Gauge["TotalMemory"] = float64(v.Total)
	liveMetrics.Gauge["FreeMemory"] = float64(v.Free)
	liveMetrics.Gauge["CPUutilization1 "] = v.UsedPercent
}

func getCounterMetrics(liveMetrics *storage.MemStorage) {
	//liveMetrics.Lock()
	liveMetrics.Counter["PollCount"] += 1
	//liveMetrics.Unlock()
}

func MetricGenerator(rm runtime.MemStats, liveMetrics *storage.MemStorage, ch chan storage.MemStorage) {
	runtime.ReadMemStats(&rm)
	getRuntimeMetrics(&rm, liveMetrics)
	getGopsMetrics(liveMetrics)
	getCounterMetrics(liveMetrics)

	ch <- *liveMetrics
}
