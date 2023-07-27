package agent

import (
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"github.com/CvitoyBamp/metricsexporter/internal/metrics"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type wants struct {
	code        int
	contentType string
}

type testMetric struct {
	metricName  string
	metricType  string
	metricValue string
}

func Test_main(t *testing.T) {

	s := &handlers.CustomServer{
		Storage: storage.CreateMemStorage(),
	}

	a := &Agent{
		Client: &http.Client{
			Timeout: 1 * time.Second,
		},
		Metrics: &metrics.Metrics{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		},
	}

	tests := []struct {
		testName   string
		testMetric testMetric
		wants      wants
	}{
		{
			testName: "Metric was successfully pushed",
			testMetric: testMetric{
				metricName:  "testGauge",
				metricType:  "gauge",
				metricValue: "1.0",
			},
			wants: wants{
				code:        http.StatusOK,
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ts := httptest.NewServer(s.MetricRouter())
			defer ts.Close()
			a.Endpoint = ts.URL[7:]
			err := a.PostMetric(tt.testMetric.metricType, tt.testMetric.metricName, tt.testMetric.metricValue)
			require.NoError(t, err)
		})
	}
}
