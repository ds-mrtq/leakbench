package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	b, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var cfg struct {
		Port     int
		TLSMode  string `json:"tls_mode"`
		Password string
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		panic(err)
	}
	if cfg.Port != 587 || cfg.TLSMode != "starttls" {
		fmt.Fprintln(os.Stderr, "TLS handshake: received plaintext SMTP greeting before STARTTLS")
		os.Exit(1)
	}
	if cfg.Password == "" {
		os.Exit(1)
	}
	fmt.Println("SMTP negotiation settings accepted")
}
