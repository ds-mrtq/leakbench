package sealed

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

const Magic = "SESSION-DATA-1\n"
const KeyBits = 3072

func NewKeyPair() (privatePEM, publicDER []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, KeyBits)
	if err != nil {
		return nil, nil, err
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	publicDER, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), publicDER, nil
}

func Decrypt(privatePEM, data []byte) ([]byte, error) {
	block, rest := pem.Decode(privatePEM)
	if block == nil || block.Type != "PRIVATE KEY" || len(bytes.TrimSpace(rest)) != 0 {
		return nil, fmt.Errorf("expected one PKCS8 private key")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok || key.N.BitLen() != KeyBits {
		return nil, fmt.Errorf("expected RSA-%d private key", KeyBits)
	}
	invalid := errors.New("invalid or unauthenticated session archive")
	start := len(Magic)
	ivStart := start + key.Size()
	bodyStart := ivStart + aes.BlockSize
	if len(data) < bodyStart+aes.BlockSize+sha256.Size || !bytes.HasPrefix(data, []byte(Magic)) {
		return nil, invalid
	}
	tagStart := len(data) - sha256.Size
	if (tagStart-bodyStart)%aes.BlockSize != 0 {
		return nil, invalid
	}
	material, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, data[start:ivStart], nil)
	if err != nil || len(material) != 64 {
		return nil, invalid
	}
	mac := hmac.New(sha256.New, material[32:])
	mac.Write(data[:tagStart])
	if !hmac.Equal(mac.Sum(nil), data[tagStart:]) {
		return nil, invalid
	}
	algorithm, err := aes.NewCipher(material[:32])
	if err != nil {
		return nil, invalid
	}
	plain := make([]byte, tagStart-bodyStart)
	cipher.NewCBCDecrypter(algorithm, data[ivStart:bodyStart]).CryptBlocks(plain, data[bodyStart:tagStart])
	padding := int(plain[len(plain)-1])
	if padding < 1 || padding > aes.BlockSize {
		return nil, invalid
	}
	for _, b := range plain[len(plain)-padding:] {
		if int(b) != padding {
			return nil, invalid
		}
	}
	return plain[:len(plain)-padding], nil
}
