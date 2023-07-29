package middlewares

import (
	"github.com/stretchr/testify/require"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type wants struct {
	contentEncoding string
}

type request struct {
	method string
	header string
}

func TestGzipWriter(t *testing.T) {

	tests := []struct {
		testName string
		request  request
		wants    wants
	}{
		{
			testName: "Contains gzip in header",
			request: request{
				method: http.MethodGet,
				header: "gzip",
			},
			wants: wants{
				contentEncoding: "gzip",
			},
		},
		{
			testName: "Uncompressed request",
			request: request{
				method: http.MethodGet,
				header: "",
			},
			wants: wants{
				contentEncoding: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ts := httptest.NewServer(MiddlewareZIP(http.DefaultServeMux))
			defer ts.Close()
			req, rerr := http.NewRequest(tt.request.method, ts.URL, nil)
			req.Header.Set("Accept-Encoding", tt.request.header)
			resp, cerr := ts.Client().Do(req)
			log.Print(resp.Header)
			require.NoError(t, rerr)
			require.NoError(t, cerr)
			defer resp.Body.Close()
			assert.Equal(t, tt.wants.contentEncoding, resp.Header.Get("Content-Encoding"))
		})
	}
}
