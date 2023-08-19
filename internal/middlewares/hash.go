package middlewares

import (
	"crypto/hmac"
	"github.com/CvitoyBamp/metricsexporter/internal/crypto"
	"io"
	"net/http"
)

func MiddlewareHash(secretKey string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

			if req.Header.Get("HashSHA256") == "" {
				h.ServeHTTP(res, req)
				return
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			data := crypto.CreateHash(body, secretKey)

			if !hmac.Equal([]byte(req.Header.Get("HashSHA256")), []byte(data)) {
				http.Error(res, "Hash isn't equal.", http.StatusBadRequest)
				return
			}

			res.Header().Set("HashSHA256", data)

			h.ServeHTTP(res, req)
		})
	}
}
