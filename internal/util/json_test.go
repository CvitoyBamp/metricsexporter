package util

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type Metric struct {
	metricValue string
	metricType  string
	metricName  string
}

func setVal(Value float64) *float64 {
	return &Value
}

func TestJsonCreator(t *testing.T) {
	tests := []struct {
		testName string
		data     Metric
		res      string
	}{
		{
			testName: "Successful parse gauge",
			data: Metric{
				metricName:  "Success",
				metricType:  "gauge",
				metricValue: "1.0",
			},
			res: "{\"id\":\"Success\",\"type\":\"gauge\",\"value\":1}",
		},
		{
			testName: "Successful parse counter",
			data: Metric{
				metricName:  "Success",
				metricType:  "counter",
				metricValue: "1",
			},
			res: "{\"id\":\"Success\",\"type\":\"counter\",\"delta\":1}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			res, err := JSONCreator(tt.data.metricValue, tt.data.metricType, tt.data.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.res, string(res))
		})
	}
}
