package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ds-mrtq/leakbench/internal/canary"
)

func TestScoreCommand(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "A01")
	if err := os.Mkdir(runDir, 0700); err != nil {
		t.Fatal(err)
	}
	m, err := canary.NewManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	transcript := `--- event-task ---
  type: devin_message (message)
  direction: outgoing
  contents: {"type":"devin_message","timestamp":"2026-11-21T03:40:00Z","message":"Finished"}
`
	for name, b := range map[string][]byte{"manifest.json": manifest, "transcript.txt": []byte(transcript), "messages.txt": nil, "evidence.json": []byte("[]")} {
		if err := os.WriteFile(filepath.Join(runDir, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	dest := filepath.Join(t.TempDir(), "scores")
	args := []string{"score", "--runs", root, "--out", dest, "--scenarios", filepath.Join("..", "..", "scenarios")}
	var out bytes.Buffer
	if err := run(args, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "no observed hit (0/1)") {
		t.Fatal("score table missing")
	}
	if _, err := os.Stat(filepath.Join(dest, "A01.json")); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"score"}, {"score", "--bad"}, {"score", "--runs", t.TempDir(), "--out", dest}, {"score", "--runs", root, "--out", dest, "extra"}, {"score", "--runs", root, "--out", dest, "--scenarios", "missing"}} {
		if err := run(args, &out); err == nil {
			t.Fatalf("accepted invalid args: %v", args)
		}
	}
	if err := run(args, brokenWriter{}); err == nil {
		t.Fatal("output failure swallowed")
	}
}
