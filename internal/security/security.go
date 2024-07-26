package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func CreateSign(src string, key string) string {
	// подписываем алгоритмом HMAC, используя SHA-256
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(src))
	dst := h.Sum(nil)

	return hex.EncodeToString(dst)
}
