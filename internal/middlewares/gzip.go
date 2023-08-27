package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// CompressWriter реализует http.ResponseWriter
//
// CompressWriter is implementation of http.ResponseWriter
type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// CompressReader реализует io.ReadCloser
//
// CompressReader is implementation of io.ReadCloser
type CompressReader struct {
	r io.ReadCloser
	z *gzip.Reader
}

// NewCompressWriter это конструктор для CompressWriter
//
// NewCompressWriter is constructor for CompressWriter
func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	zw := gzip.NewWriter(w)
	return &CompressWriter{w, zw}
}

// Header реализует http.ResponseWriter.Header
//
// Header is implementation of http.ResponseWriter.Header
func (cw *CompressWriter) Header() http.Header {
	return cw.w.Header()
}

// Write реализует http.ResponseWriter.Write
//
// Write is implementation of http.ResponseWriter.Write
func (cw *CompressWriter) Write(b []byte) (int, error) {
	return cw.zw.Write(b)
}

// WriteHeader реализует http.ResponseWriter.WriteHeader
//
// WriteHeader is implementation of http.ResponseWriter.WriteHeader
func (cw *CompressWriter) WriteHeader(statusCode int) {
	cw.w.WriteHeader(statusCode)
	if statusCode > 199 && statusCode < 300 {
		cw.w.Header().Set("Content-Encoding", "gzip")
	}
}

// Close закрывает gzip.Writer
//
// Close is closes gzip.Writer
func (cw *CompressWriter) Close() error {
	return cw.zw.Close()
}

// NewCompressReader это конструктор для CompressReader
//
// NewCompressReader is constructor for CompressReader
func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	z, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &CompressReader{r, z}, nil
}

// Read реализует io.ReadCloser.Read
//
// Read is implementation of io.ReadCloser.Read
func (cr *CompressReader) Read(b []byte) (int, error) {
	return cr.z.Read(b)
}

// Close реализует io.ReadCloser.Close
//
// Close is implementation of io.ReadCloser.Close
func (cr *CompressReader) Close() error {
	if err := cr.z.Close(); err != nil {
		return err
	}
	return cr.r.Close()
}

// ZipMiddleware это middleware, который распаковывает запросы и сжимает ответы сервера с помощью gzip.
//
// ZipMiddleware is middleware that unpacks requests and compresses server responses using gzip.
func MiddlewareZIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		contentEncoding := r.Header.Get("Content-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		sendGzip := strings.Contains(contentEncoding, "gzip")
		// this section for ordinary request
		if !supportGzip && !sendGzip {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
			return
		}
		// this section for request with accept-encoding: gzip
		if supportGzip && !sendGzip {
			originWriter := w
			compressedWriter := NewCompressWriter(w)

			originWriter = compressedWriter
			originWriter.Header().Set("Content-Encoding", "gzip")
			//defer compressedWriter.Close()
			defer func() {
				err := compressedWriter.Close()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(originWriter, r)
		}
		// this section for request with content-encoding: gzip
		if sendGzip {
			//originWriter := w
			//compressedWriter := NewCompressWriter(w)
			//originWriter = compressedWriter
			originWriter := NewCompressWriter(w)
			originWriter.Header().Set("Content-Encoding", "gzip")
			//defer compressedWriter.Close()
			defer originWriter.Close()
			compressedReader, err := NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			r.Body = compressedReader
			defer compressedReader.Close()

			next.ServeHTTP(originWriter, r)

		}
	})

}
