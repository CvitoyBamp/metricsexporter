package metrics

import (
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
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
		metricsTest *storage.MemStorage
		wants       wants
	}{
		{
			testName:    "Check metrics exist",
			memStats:    &runtime.MemStats{},
			metricsTest: storage.CreateMemStorage(),
			wants: wants{
				hasCustomCounterMetric: true,
				hasCustomGaugeMetric:   true,
				hasRuntimeMetric:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			m := make(chan storage.MemStorage, 3)
			MetricGenerator(*tt.memStats, m)
			_, pc := tt.metricsTest.Counter["PollCount"]
			assert.Equal(t, tt.wants.hasCustomCounterMetric, pc)
			_, ta := tt.metricsTest.Gauge["TotalAlloc"]
			assert.Equal(t, tt.wants.hasRuntimeMetric, ta)
			_, rv := tt.metricsTest.Gauge["RandomValue"]
			assert.Equal(t, tt.wants.hasCustomGaugeMetric, rv)
		})
	}
}
