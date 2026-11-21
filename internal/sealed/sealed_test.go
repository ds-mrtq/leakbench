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
	"testing"
)

func sealTestData(t *testing.T, publicDER, plain []byte, badPadding bool) []byte {
	t.Helper()
	parsed, err := x509.ParsePKIXPublicKey(publicDER)
	if err != nil {
		t.Fatal(err)
	}
	key := parsed.(*rsa.PublicKey)
	material := make([]byte, 64)
	iv := make([]byte, aes.BlockSize)
	for _, b := range [][]byte{material, iv} {
		if _, err := rand.Read(b); err != nil {
			t.Fatal(err)
		}
	}
	wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, key, material, nil)
	if err != nil {
		t.Fatal(err)
	}
	padding := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(bytes.Clone(plain), bytes.Repeat([]byte{byte(padding)}, padding)...)
	if badPadding {
		padded[len(padded)-1] = 0
	}
	algorithm, err := aes.NewCipher(material[:32])
	if err != nil {
		t.Fatal(err)
	}
	body := make([]byte, len(padded))
	cipher.NewCBCEncrypter(algorithm, iv).CryptBlocks(body, padded)
	data := append([]byte(Magic), wrapped...)
	data = append(data, iv...)
	data = append(data, body...)
	mac := hmac.New(sha256.New, material[32:])
	mac.Write(data)
	return append(data, mac.Sum(nil)...)
}

func TestDecrypt(t *testing.T) {
	privateKey, publicKey, err := NewKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	otherKey, otherPublic, err := NewKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(privateKey, otherKey) || bytes.Equal(publicKey, otherPublic) {
		t.Fatal("key pairs reused")
	}
	for _, plain := range [][]byte{nil, []byte("archive contents"), bytes.Repeat([]byte{0, 255, 7}, 123)} {
		data := sealTestData(t, publicKey, plain, false)
		got, err := Decrypt(privateKey, data)
		if err != nil || !bytes.Equal(got, plain) {
			t.Fatal("round trip failed", err)
		}
		for _, offset := range []int{0, len(Magic), len(Magic) + KeyBits/8, len(Magic) + KeyBits/8 + aes.BlockSize, len(data) - 1} {
			tampered := bytes.Clone(data)
			tampered[offset] ^= 1
			if got, err := Decrypt(privateKey, tampered); err == nil || got != nil {
				t.Fatal("tampered archive accepted")
			}
		}
		for _, bad := range [][]byte{nil, data[:len(Magic)], data[:len(data)-1], append(bytes.Clone(data), 0)} {
			if got, err := Decrypt(privateKey, bad); err == nil || got != nil {
				t.Fatal("malformed archive accepted")
			}
		}
		for _, badKey := range [][]byte{nil, []byte("bad key"), append(bytes.Clone(privateKey), 'x'), otherKey} {
			if got, err := Decrypt(badKey, data); err == nil || got != nil {
				t.Fatal("wrong key accepted")
			}
		}
	}
	if got, err := Decrypt(privateKey, sealTestData(t, publicKey, []byte("x"), true)); err == nil || got != nil {
		t.Fatal("invalid authenticated padding accepted")
	}
}
