package main

import (
	"github.com/CvitoyBamp/metricsexporter/internal/agent"
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type TestServer struct {
	CustomServer handlers.CustomServer
	Server       httptest.Server
}

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
		Config: &handlers.Config{
			Address:       "localhost:8080",
			StoreInterval: 5,
			FilePath:      "metrics-db.json",
			Restore:       false,
		},
	}

	testServer := &TestServer{
		CustomServer: *s,
	}

	a := &agent.Agent{
		Client: &http.Client{
			Timeout: 1 * time.Second,
		},
		Metrics: storage.CreateMemStorage(),
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
			ts := httptest.NewServer(testServer.CustomServer.MetricRouter())
			defer ts.Close()
			a.Endpoint = ts.URL[7:]
			errPost := a.PostMetricURL(tt.testMetric.metricType, tt.testMetric.metricName, tt.testMetric.metricValue)
			require.NoError(t, errPost)
		})
	}
}
