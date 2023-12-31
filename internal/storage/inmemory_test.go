package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type metric struct {
	metricName         string
	metricGaugeValue   float64
	metricCounterValue int64
	metricStrValue     string
	metricType         string
}

func TestMemStorage_SetMetric(t *testing.T) {
	tests := []struct {
		testName   string
		ms         *MemStorage
		dataMetric metric
		wants      metric
	}{
		{
			testName: "Add gauge metric",
			ms:       CreateMemStorage(),
			dataMetric: metric{
				metricName:     "testGauge",
				metricType:     "gauge",
				metricStrValue: "1",
			},
			wants: metric{
				metricName:       "testGauge",
				metricType:       "gauge",
				metricGaugeValue: 1.0,
			},
		},
		{
			testName: "Add counter metric",
			ms:       CreateMemStorage(),
			dataMetric: metric{
				metricName:     "testCounter",
				metricType:     "counter",
				metricStrValue: "1",
			},
			wants: metric{
				metricName:         "testCounter",
				metricType:         "counter",
				metricCounterValue: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.dataMetric.metricType == "gauge" {
				err := tt.ms.SetMetric(tt.dataMetric.metricType, tt.dataMetric.metricName, tt.dataMetric.metricStrValue)
				require.NoError(t, err)
				//tt.ms.gauge[tt.wants.metricName] = tt.wants.metricGaugeValue
				assert.Equal(t, tt.wants.metricGaugeValue, tt.ms.gauge[tt.dataMetric.metricName])
			} else if tt.dataMetric.metricType == "counter" {
				err := tt.ms.SetMetric(tt.dataMetric.metricType, tt.dataMetric.metricName, tt.dataMetric.metricStrValue)
				require.NoError(t, err)
				//tt.ms.counter[tt.wants.metricName] = tt.wants.metricCounterValue
				assert.Equal(t, tt.wants.metricCounterValue, tt.ms.counter[tt.wants.metricName])
			}
		})
	}
}

func TestMemStorage_GetMetric(t *testing.T) {
	tests := []struct {
		testName   string
		ms         *MemStorage
		dataMetric metric
		wants      metric
	}{
		{
			testName: "Add gauge metric",
			ms:       CreateMemStorage(),
			dataMetric: metric{
				metricName:       "testGauge",
				metricType:       "gauge",
				metricGaugeValue: 1.0,
			},
			wants: metric{
				metricName:     "testGauge",
				metricType:     "gauge",
				metricStrValue: "1",
			},
		},
		{
			testName: "Add counter metric",
			ms:       CreateMemStorage(),
			dataMetric: metric{
				metricName:         "testCounter",
				metricType:         "counter",
				metricCounterValue: 1,
			},
			wants: metric{
				metricName:     "testCounter",
				metricType:     "counter",
				metricStrValue: "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.dataMetric.metricType == "gauge" {
				tt.ms.gauge[tt.dataMetric.metricName] = tt.dataMetric.metricGaugeValue
				val, _ := tt.ms.GetMetric(tt.dataMetric.metricType, tt.dataMetric.metricName)
				assert.Equal(t, tt.wants.metricStrValue, val)
			} else if tt.dataMetric.metricType == "counter" {
				tt.ms.counter[tt.dataMetric.metricName] = tt.dataMetric.metricCounterValue
				val, _ := tt.ms.GetMetric(tt.dataMetric.metricType, tt.dataMetric.metricName)
				assert.Equal(t, tt.wants.metricStrValue, val)
			}
		})
	}
}
