package middlewares

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

type CompressReader struct {
	r io.ReadCloser
	z *gzip.Reader
}

func NewCompressWriter(res http.ResponseWriter) *GzipWriter {
	zw := gzip.NewWriter(res)
	return &GzipWriter{res, zw}
}

func (gw *GzipWriter) Header() http.Header {
	return gw.w.Header()
}

func (gw *GzipWriter) Write(b []byte) (int, error) {
	return gw.zw.Write(b)
}

func (gw *GzipWriter) WriteHeader(statusCode int) {
	gw.w.WriteHeader(statusCode)
	if statusCode > 199 && statusCode < 300 {
		gw.w.Header().Set("Content-Encoding", "gzip")
	}
}

func (gw *GzipWriter) Close() error {
	return gw.zw.Close()
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	z, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &CompressReader{r, z}, nil
}

func (cr *CompressReader) Read(b []byte) (int, error) {
	return cr.z.Read(b)
}

func (cr *CompressReader) Close() error {
	if err := cr.z.Close(); err != nil {
		return err
	}
	return cr.r.Close()
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

func MiddlewareZIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		acceptEncoding := req.Header.Get("Accept-Encoding")
		contentEncoding := req.Header.Get("Content-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		sendGzip := strings.Contains(contentEncoding, "gzip")

		if !supportGzip && !sendGzip {
			h.ServeHTTP(res, req)
			return
		}

		if supportGzip && !sendGzip {
			originWriter := res
			compressedWriter := NewCompressWriter(res)

			originWriter = compressedWriter
			originWriter.Header().Set("Content-Encoding", "gzip")
			defer func() {
				err := compressedWriter.Close()
				if err != nil {
					res.WriteHeader(http.StatusInternalServerError)
				}
			}()
			h.ServeHTTP(originWriter, req)
		}

		if sendGzip {
			originWriter := NewCompressWriter(res)
			originWriter.Header().Set("Content-Encoding", "gzip")
			defer originWriter.Close()
			compressedReader, err := NewCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
			}
			req.Body = compressedReader
			defer compressedReader.Close()

			h.ServeHTTP(originWriter, req)

		}
	})
}
