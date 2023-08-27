package middlewares

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
)

func checkDataHash(checkSum string, secretKey string, data []byte) (bool, error) {
	if secretKey == "" {
		return false, fmt.Errorf("key is null")
	}
	requestCheckSum := sha256.Sum256(data)
	controlCheckSum := fmt.Sprintf("%x", requestCheckSum)
	if checkSum != controlCheckSum {
		fmt.Println("wrong checksum")
		fmt.Println(checkSum)
		fmt.Println(controlCheckSum)
		return false, nil
	}
	return true, nil
}

func GetHash(secretKey string, data []byte) ([32]byte, error) {
	if secretKey == "" {
		return [32]byte{}, fmt.Errorf("key is null")
	}
	checkSum := sha256.Sum256(data)
	return checkSum, nil
}

func MiddlewareHash(secretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			var buf bytes.Buffer
			_, err := buf.ReadFrom(req.Body)
			if err != nil {
				http.Error(res, "Bad request", http.StatusBadRequest)
				return
			}
			checkSum := req.Header.Get("HashSHA256")
			if checkSum != "" && secretKey != "" {
				ok, errHash := checkDataHash(checkSum, secretKey, buf.Bytes())
				if errHash != nil {
					http.Error(res, "Bad request", http.StatusBadRequest)
					return
				}
				if !ok {
					http.Error(res, "Bad request", http.StatusBadRequest)
					return
				}
			}
			req.Body = io.NopCloser(&buf)
			capture := &responseCapture{res: res}
			next.ServeHTTP(capture, req)
			hash, errGH := GetHash(secretKey, capture.body)
			if errGH != nil {
				http.Error(res, "Bad request", http.StatusBadRequest)
				return
			}
			res.Header().Set("HashSHA256", fmt.Sprintf("%x", hash))
		})
	}
}

type responseCapture struct {
	res  http.ResponseWriter
	body []byte
}

func (rc *responseCapture) Header() http.Header {
	return rc.res.Header()
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body = append(rc.body, b...)
	return rc.res.Write(b)
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.res.WriteHeader(statusCode)
}
