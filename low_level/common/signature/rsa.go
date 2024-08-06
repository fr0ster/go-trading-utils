package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"log"
)

type SignRSA struct {
	privateKey rsa.PrivateKey
	publicKey  rsa.PublicKey
}

// Функція для створення підпису RSA
func (sign *SignRSA) CreateSignature(queryString string) string {
	hashed := sha256.Sum256([]byte(queryString))
	signature, err := rsa.SignPKCS1v15(rand.Reader, &sign.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		log.Fatalf("Error signing query string: %v", err)
	}
	return base64.StdEncoding.EncodeToString(signature)
}

func (sign *SignRSA) GetAPIKey() string {
	return sign.publicKey.N.String()
}

func NewSignRSA(privateKey rsa.PrivateKey, publicKey rsa.PublicKey) *SignRSA {
	return &SignRSA{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}
