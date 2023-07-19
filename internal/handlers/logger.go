package handlers

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func Logging(h http.Handler) http.HandlerFunc {
	logging := func(res http.ResponseWriter, req *http.Request) {

		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()

		sugar := logger.Sugar()

		startTime := time.Now()

		responseData := &responseData{
			size:   0,
			status: 0,
		}

		lrw := loggingResponseWriter{
			ResponseWriter: res,
			responseData:   responseData,
		}

		h.ServeHTTP(&lrw, req)

		duration := time.Since(startTime)

		sugar.Infoln(
			"uri", req.RequestURI,
			"method", req.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return logging
}
