// Package canary generates synthetic secrets, never provider credentials.
package canary

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
)

type Kind string

const (
	Password  Kind = "password"
	APIToken  Kind = "api_token"
	HexKey    Kind = "hex_key"
	JWTSecret Kind = "jwt_secret"
)

type Canary struct {
	Name  string `json:"name"`
	Kind  Kind   `json:"kind"`
	Value string `json:"value"`
}

// Control is generated independently and must never be planted in artifacts.
type Manifest struct {
	Canaries []Canary `json:"canaries"`
	Control  Canary   `json:"control"`
}

func randomChars(alphabet string, n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		v, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[v.Int64()]
	}
	return string(b), nil
}

func Generate(name string, kind Kind) (Canary, error) {
	c := Canary{Name: name, Kind: kind}
	const lower = "abcdefghijklmnopqrstuvwxyz"
	const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const digits = "0123456789"
	const symbols = "!@%+"
	var err error
	switch kind {
	case Password:
		// Rejection sampling guarantees each character class without fixed positions.
		for {
			c.Value, err = randomChars(lower+upper+digits+symbols, 20)
			if err != nil {
				break
			}
			classes := [4]bool{}
			for _, ch := range []byte(c.Value) {
				for i, alphabet := range []string{lower, upper, digits, symbols} {
					for _, candidate := range []byte(alphabet) {
						if ch == candidate {
							classes[i] = true
						}
					}
				}
			}
			if classes[0] && classes[1] && classes[2] && classes[3] {
				break
			}
		}
	case APIToken:
		c.Value, err = randomChars(lower+upper+digits, 36)
	case HexKey, JWTSecret:
		n := 16
		if kind == JWTSecret {
			n = 32
		}
		b := make([]byte, n)
		_, err = rand.Read(b)
		if kind == HexKey {
			c.Value = hex.EncodeToString(b)
		} else {
			c.Value = base64.RawURLEncoding.EncodeToString(b)
		}
	default:
		err = fmt.Errorf("unsupported canary kind %q", kind)
	}
	if err != nil {
		return Canary{}, fmt.Errorf("generate %s: %w", name, err)
	}
	return c, nil
}

// NewManifest creates one canary per kind and a separate, unplanted control.
func NewManifest() (Manifest, error) {
	m := Manifest{Canaries: make([]Canary, 0, 4)}
	for _, kind := range []Kind{Password, APIToken, HexKey, JWTSecret} {
		c, err := Generate(string(kind), kind)
		if err != nil {
			return Manifest{}, err
		}
		m.Canaries = append(m.Canaries, c)
	}
	var err error
	m.Control, err = Generate("control", APIToken)
	return m, err
}
