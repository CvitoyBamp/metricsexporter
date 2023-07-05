package handlers

import (
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type wants struct {
	code        int
	contentType string
}

type request struct {
	url    string
	method string
}

func TestMetricCreatorHandler(t *testing.T) {

	s := &Server{
		Storage: storage.CreateMemStorage(),
	}

	tests := []struct {
		testName string
		request  request
		wants    wants
	}{
		{
			testName: "Metric was successfully added",
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
			testName: "Not a POST-method",
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
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			r := httptest.NewRequest(tt.request.method, tt.request.url, nil)
			w := httptest.NewRecorder()
			s.MetricCreatorHandler(w, r)

			assert.Equal(t, tt.wants.code, w.Code)
			assert.Equal(t, tt.wants.contentType, w.Header().Get("Content-Type"))

		})
	}
}
