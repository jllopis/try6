package try6

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/jllopis/try6/log"
)

// NewKey genera una pareja de claves RSA de 2048 bits. Las clave privada se codifica como PKCS1 y la pública como PKIX.
// Ambas en formato PEM.
func NewKey(uid string) *Key {
	if uid == "" {
		log.LogE("new key needs an account", "pkg", "try6", "func", "NewKey(string) *Key")
		return nil
	}
	k := newKey()
	k.AccountID = uid
	return k
}

// newKey realiza la generación y codificación de las claves RSA en codificación PEM.
func newKey() *Key {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.LogE("failed to generate private key", "pkg", "try6", "func", "NewKey(string) *Key", "error", err.Error())
		return nil
	}
	privPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	pubKeyPKIX, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.LogE("failed to generate DER public key", "pkg", "try6", "func", "NewKey(string) *Key", "error", err.Error())
		return nil
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyPKIX,
	})

	return &Key{
		PubKey:  pubPEM,
		PrivKey: privPEM,
	}
}
