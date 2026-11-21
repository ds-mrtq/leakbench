package main

import (
	"bytes"
	"compress/gzip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecryptOpenSSLRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("openssl unavailable")
	}
	root := filepath.Join(t.TempDir(), "run")
	if err := run([]string{"render", "--scenario", filepath.Join("..", "..", "scenarios", "A01"), "--out", root}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", filepath.Join(root, "collect.sh"))
	cmd.Env = []string{"HOME=" + root, "PATH=" + os.Getenv("PATH")}
	output, err := cmd.CombinedOutput()
	input := filepath.Join(root, "session-archive.bin")
	if err != nil || strings.TrimSpace(string(output)) != input {
		t.Fatalf("collection failed: %v %s", err, output)
	}
	key := filepath.Join(root, "collection-key.pem")
	for i, args := range [][]string{
		{"decrypt", "--key", key, input, "--out", filepath.Join(root, "first.tar.gz")},
		{"decrypt", "--key", key, "--out", filepath.Join(root, "second.tar.gz"), input},
	} {
		var out bytes.Buffer
		if err := run(args, &out); err != nil {
			t.Fatal(err)
		}
		path := strings.TrimSpace(out.String())
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		gz, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		gz.Close()
		f.Close()
		if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
			t.Fatal("plaintext permissions", err)
		}
		if err := run(args, &bytes.Buffer{}); err == nil {
			t.Fatal("overwrote plaintext")
		}
		if i == 0 {
			if err := run([]string{"decrypt", "--key", key, input, "--out", path + "writer"}, brokenWriter{}); err == nil {
				t.Fatal("missing writer error")
			}
		}
	}
	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-40] ^= 1
	tampered := filepath.Join(root, "tampered.bin")
	if err := os.WriteFile(tampered, data, 0600); err != nil {
		t.Fatal(err)
	}
	failedOut := filepath.Join(root, "failed.tar.gz")
	if err := run([]string{"decrypt", "--key", key, tampered, "--out", failedOut}, &bytes.Buffer{}); err == nil {
		t.Fatal("tampered archive accepted")
	}
	if _, err := os.Stat(failedOut); !os.IsNotExist(err) {
		t.Fatal("plaintext emitted before authentication")
	}
	for _, args := range [][]string{
		{"decrypt"}, {"decrypt", "--bad"}, {"decrypt", "--key", key},
		{"decrypt", "--key", "missing", input, "--out", failedOut},
		{"decrypt", "--key", key, "missing", "--out", failedOut},
		{"decrypt", "--key", key, input, input, "--out", failedOut},
	} {
		if err := run(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("accepted invalid arguments: %v", args)
		}
	}
}
