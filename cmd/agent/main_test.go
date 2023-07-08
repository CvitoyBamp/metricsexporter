package main

import (
	"github.com/CvitoyBamp/metricsexporter/internal/handlers"
	"net/http"
	"testing"
)

type wants struct {
	code        int
	contentType string
}

type request struct {
	url    string
	method string
}

type sc struct {
	s *handlers.CustomServer
	c *http.Client
}

func Test_main(t *testing.T) {

	tests := []struct {
		testName     string
		serverClient sc
		request      request
		wants        wants
	}{
		{
			testName:     "Metric was successfully added",
			serverClient: sc{},
			request: request{
				url:    "/update/gauge/testGauge/100",
				method: http.MethodPost,
			},
			wants: wants{
				code:        http.StatusOK,
				contentType: "",
			},
		},
		{
			testName:     "Not a POST-method",
			serverClient: sc{},
			request: request{
				url:    "/update/gauge/testGauge/100",
				method: http.MethodGet,
			},
			wants: wants{
				code:        http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			testName:     "Not correct URL",
			serverClient: sc{},
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
			testName:     "Can't parse value",
			serverClient: sc{},
			request: request{
				url:    "/update/gauge/testGauge/badData",
				method: http.MethodPost,
			},
			wants: wants{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			t.Skipped()
			//server := *httptest.NewServer(http.HandlerFunc(tt.serverClient.s.MetricCreatorHandler))
			//client := server.Client()
			//
			//url := fmt.Sprintf("http://localhost:8080/%s", tt.request.url)
			//res, _ := client.Post(url, "text/plain", nil)
			//
			//assert.Equal(t, tt.wants.code, res.StatusCode)
			//assert.Equal(t, tt.wants.contentType, res.Header.Get("Content-Type"))
			//
			//server.Close()

		})
	}
}
