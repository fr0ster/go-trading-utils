package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
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

func NewSignRSA(publicKey PublicKey, privateKey SecretKey) (sign *SignRSA, err error) {
	private, err := loadPrivateKeyFromPEM(string(privateKey))
	if err != nil {
		return
	}
	public, err := loadPublicKeyFromPEM(string(publicKey))
	if err != nil {
		return
	}

	sign = &SignRSA{
		privateKey: *private,
		publicKey:  *public,
	}
	return
}

// Функція для завантаження приватного ключа з PEM рядка
func loadPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// Функція для завантаження публічного ключа з PEM рядка
func loadPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}
