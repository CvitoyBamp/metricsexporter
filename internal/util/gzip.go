package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func MiddlewareZIP(h http.Handler) http.Handler {
	zip := func(res http.ResponseWriter, req *http.Request) {

		//if !strings.Contains(req.Header.Get("Content-Type"), "application/json") &&
		//	!strings.Contains(req.Header.Get("Content-Type"), "text/html") {
		//	h.ServeHTTP(res, req)
		//	return
		//}

		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(res, req)
			return
		}

		gz, err := gzip.NewWriterLevel(res, gzip.BestSpeed)
		if err != nil {
			io.WriteString(res, err.Error())
			return
		}
		defer gz.Close()

		//if strings.Contains(req.Header.Get("Content-Encoding"), "gzip") {
		//	req.Body, err = gzip.NewReader(req.Body)
		//	if err != nil {
		//		http.Error(res, err.Error(), http.StatusInternalServerError)
		//		return
		//	}
		//}

		res.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzipWriter{ResponseWriter: res, Writer: gz}, req)
	}
	return http.HandlerFunc(zip)
}

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %v", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress buffer: %v", err)
	}
	w.Close()
	return b.Bytes(), nil
}
