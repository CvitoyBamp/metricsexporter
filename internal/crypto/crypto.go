package crypto

import (
	"crypto/sha256"
)

func CreateHash(data []byte, secretKey string) [32]byte {
	if secretKey == "" {
		return [32]byte{}
	}
	checkSum := sha256.Sum256(data)

	return checkSum
}
