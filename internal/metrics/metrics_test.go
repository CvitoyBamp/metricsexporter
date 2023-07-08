package metrics

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

type wants struct {
	hasRuntimeMetric       bool
	hasCustomGaugeMetric   bool
	hasCustomCounterMetric bool
}

func TestMetrics_MetricGenerator(t *testing.T) {

	tests := []struct {
		testName    string
		memStats    *runtime.MemStats
		metricsTest *Metrics
		wants       wants
	}{
		{
			testName: "Check metrics exist",
			memStats: &runtime.MemStats{},
			metricsTest: &Metrics{
				Gauge:   make(map[string]float64),
				Counter: make(map[string]int64),
			},
			wants: wants{
				hasCustomCounterMetric: true,
				hasCustomGaugeMetric:   true,
				hasRuntimeMetric:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tt.metricsTest.MetricGenerator(*tt.memStats)
			_, pc := tt.metricsTest.Counter["PollCount"]
			assert.Equal(t, tt.wants.hasCustomCounterMetric, pc)
			_, ta := tt.metricsTest.Gauge["TotalAlloc"]
			assert.Equal(t, tt.wants.hasRuntimeMetric, ta)
			_, rv := tt.metricsTest.Gauge["RandomValue"]
			assert.Equal(t, tt.wants.hasCustomGaugeMetric, rv)
		})
	}
}
