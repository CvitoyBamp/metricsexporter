package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func CreateHash(data []byte, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(data)
	dst := h.Sum(nil)
	return hex.EncodeToString(dst)
}
