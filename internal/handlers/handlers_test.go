package handlers

import (
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestServer struct {
	CustomServer CustomServer
	Server       httptest.Server
}

type wants struct {
	code        int
	contentType string
	value       string
}

type request struct {
	url    string
	method string
}

func TestMetricCreatorHandler(t *testing.T) {

	s := &CustomServer{
		Storage: storage.CreateMemStorage(),
		Config: &Config{
			StoreInterval: 5,
			FilePath:      "metrics-db.json",
			Restore:       false,
		},
	}

	testServer := &TestServer{
		CustomServer: *s,
	}

	tests := []struct {
		testName string
		request  request
		wants    wants
	}{
		{
			testName: "Metric was successfully added",
			request: request{
				url:    "/update/gauge/testGauge/100.1",
				method: http.MethodPost,
			},
			wants: wants{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			testName: "Metric was successfully added",
			request: request{
				url:    "/update/counter/testCounter/100",
				method: http.MethodPost,
			},
			wants: wants{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			testName: "Not correct URL for a POST-method",
			request: request{
				url:    "/update/gauge/testGauge/100.1",
				method: http.MethodGet,
			},
			wants: wants{
				code:        http.StatusMethodNotAllowed,
				contentType: "",
			},
		},
		{
			testName: "Not correct URL",
			request: request{
				url:    "/update/gauge/",
				method: http.MethodPost,
			},
			wants: wants{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			testName: "Can't parse value",
			request: request{
				url:    "/update/gauge/testGauge/badData",
				method: http.MethodPost,
			},
			wants: wants{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			testName: "Get metric by Name",
			request: request{
				url:    "/value/gauge/testGauge",
				method: http.MethodGet,
			},
			wants: wants{
				code:        http.StatusOK,
				contentType: "",
				value:       "100.1",
			},
		},
		{
			testName: "Get unexist metric",
			request: request{
				url:    "/value/gauge/UnExistMetric",
				method: http.MethodGet,
			},
			wants: wants{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			testName: "Get metrics",
			request: request{
				url:    "/",
				method: http.MethodGet,
			},
			wants: wants{
				code:        http.StatusOK,
				contentType: "text/html",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ts := httptest.NewServer(testServer.CustomServer.MetricRouter())
			defer ts.Close()
			req, rerr := http.NewRequest(tt.request.method, ts.URL+tt.request.url, nil)
			resp, cerr := ts.Client().Do(req)
			require.NoError(t, rerr)
			require.NoError(t, cerr)
			defer resp.Body.Close()
			assert.Equal(t, tt.wants.code, resp.StatusCode)
			assert.Equal(t, tt.wants.contentType, resp.Header.Get("Content-Type"))

			if tt.wants.value != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.wants.value, string(body))
			}
		})
	}
}
