package json

import (
	"bytes"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type Metric struct {
	metricValue string
	metricType  string
	metricName  string
}

func TestCreator(t *testing.T) {
	tests := []struct {
		testName string
		data     Metric
		res      string
	}{
		{
			testName: "Successful create gauge",
			data: Metric{
				metricName:  "Success",
				metricType:  "gauge",
				metricValue: "1.0",
			},
			res: "{\"id\":\"Success\",\"type\":\"gauge\",\"value\":1}",
		},
		{
			testName: "Successful create counter",
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
			res, err := Creator(tt.data.metricValue, tt.data.metricType, tt.data.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.res, string(res))
		})
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		testName string
		req      string
		data     Metric
	}{
		{
			testName: "Successful parse gauge",
			req:      "{\"id\":\"Success\",\"type\":\"gauge\",\"value\":1}",
			data: Metric{
				metricName:  "Success",
				metricType:  "gauge",
				metricValue: "1.0",
			},
		},
		{
			testName: "Successful parse counter",
			req:      "{\"id\":\"Success\",\"type\":\"counter\",\"delta\":1}",
			data: Metric{
				metricName:  "Success",
				metricType:  "counter",
				metricValue: "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			request, errReq := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.req)))
			require.NoError(t, errReq)
			_ = func(res http.ResponseWriter, req *http.Request) {
				assert.Equal(t, tt.data, Parser(res, request))
			}
		})
	}
}

func TestMetricConverter(t *testing.T) {

	ms := storage.CreateMemStorage()

	tests := []struct {
		testName string
		resp     string
		data     Metric
	}{
		{
			testName: "Successful convert gauge",
			resp:     "[{\"id\":\"Success\",\"type\":\"gauge\",\"value\":1}]",
			data: Metric{
				metricName:  "Success",
				metricType:  "gauge",
				metricValue: "1.0",
			},
		},
		{
			testName: "Successful convert counter",
			resp:     "[{\"id\":\"Success\",\"type\":\"counter\",\"delta\":5}]",
			data: Metric{
				metricName:  "Success",
				metricType:  "counter",
				metricValue: "5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			errSet := ms.SetMetric(tt.data.metricType, tt.data.metricName, tt.data.metricValue)
			require.NoError(t, errSet)
			b, errConv := MetricConverter(ms)
			require.NoError(t, errConv)
			assert.Equal(t, tt.resp, string(b))
			errDel := ms.DeleteMetric(tt.data.metricType, tt.data.metricName)
			require.NoError(t, errDel)
		})
	}
}
