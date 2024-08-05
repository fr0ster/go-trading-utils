package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Функція для створення підпису
func CreateSignatureHMAC(apiSecret, queryString string) string {
	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

type Sign struct {
	apiKey    string
	apiSecret string
}

func (s *Sign) GetAPIKey() string {
	return s.apiKey
}

func (s *Sign) GetAPISecret() string {
	return s.apiSecret
}

func NewSign(apiKey, apiSecret string) *Sign {
	return &Sign{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}
