package canary

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	seen := map[string]bool{}
	for _, kind := range []Kind{Password, APIToken, HexKey, JWTSecret} {
		t.Run(string(kind), func(t *testing.T) {
			for range 100 {
				c, err := Generate("test", kind)
				if err != nil {
					t.Fatal(err)
				}
				if c.Name != "test" || c.Kind != kind || len(c.Value) < 16 || seen[c.Value] {
					t.Fatalf("invalid or duplicate canary: %v", c)
				}
				seen[c.Value] = true
				switch kind {
				case Password:
					if len(c.Value) != 20 {
						t.Fatal("password length")
					}
					for _, alphabet := range []string{"abcdefghijklmnopqrstuvwxyz", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "0123456789", "!@%+"} {
						if !strings.ContainsAny(c.Value, alphabet) {
							t.Fatal("missing character class")
						}
					}
				case APIToken:
					if len(c.Value) != 36 {
						t.Fatal("token length")
					}
					for _, ch := range c.Value {
						if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", ch) {
							t.Fatal("non-base62 token")
						}
					}
				case HexKey:
					if b, err := hex.DecodeString(c.Value); err != nil || len(b) != 16 {
						t.Fatal("invalid hex key")
					}
				case JWTSecret:
					if b, err := base64.RawURLEncoding.DecodeString(c.Value); err != nil || len(b) != 32 || len(c.Value) != 43 {
						t.Fatal("invalid JWT secret")
					}
				}
			}
		})
	}
}

func TestManifest(t *testing.T) {
	m, err := NewManifest()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Canaries) != 4 || m.Control.Name != "control" {
		t.Fatal("missing canaries")
	}
	for _, c := range m.Canaries {
		if c.Value == m.Control.Value {
			t.Fatal("control reused")
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var round Manifest
	if err = json.Unmarshal(b, &round); err != nil {
		t.Fatal(err)
	}
	if round.Control != m.Control || len(round.Canaries) != 4 {
		t.Fatal("round trip failed")
	}
}

func TestUnsupportedKind(t *testing.T) {
	if _, err := Generate("test", Kind("invalid")); err == nil {
		t.Fatal("expected error")
	}
}
