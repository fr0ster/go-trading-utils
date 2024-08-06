package signature

import (
	"crypto/ed25519"
	"encoding/base64"
)

type SignEd25519 struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// Функція для створення підпису Ed25519
func (sign *SignEd25519) CreateSignature(queryString string) string {
	signature := ed25519.Sign(sign.privateKey, []byte(queryString))
	return base64.StdEncoding.EncodeToString(signature)
}

func (sign *SignEd25519) GetAPIKey() string {
	return string(sign.publicKey)
}

func NewSignEd25519(publicKey PublicKey, privateKey SecretKey) *SignEd25519 {
	return &SignEd25519{
		privateKey: ed25519.PrivateKey(privateKey),
		publicKey:  ed25519.PublicKey(publicKey),
	}
}
