package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ds-mrtq/leakbench/internal/canary"
	"github.com/ds-mrtq/leakbench/internal/detect"
)

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestDetect(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "manifest.json")
	artifact := filepath.Join(dir, "artifact")
	m, err := canary.NewManifest()
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(manifest, b, 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, text string
		want       int
	}{{"hit", m.Canaries[0].Value, 1}, {"empty", "", 0}, {"no hits", "unrelated", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(artifact, []byte(tc.text), 0600); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := run([]string{"detect", "--manifest", manifest, artifact}, &out); err != nil {
				t.Fatal(err)
			}
			var result detect.Result
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if len(result.Hits) != tc.want {
				t.Fatalf("unexpected JSON: %s", out.String())
			}
		})
	}
	if err := run([]string{"detect", "--manifest", manifest, artifact}, brokenWriter{}); err == nil {
		t.Fatal("expected output error")
	}
	for _, args := range [][]string{nil, {"unknown"}, {"detect"}, {"detect", "--bad"}, {"detect", "--manifest", manifest}, {"detect", "--manifest", manifest, "missing-file"}, {"detect", "--manifest", "missing-manifest", artifact}} {
		if err := run(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("expected error for %v", args)
		}
	}
	for _, text := range []string{"{", "null", "{}", `{"canaries":[{"name":"bad","value":"short"}]}`} {
		if err := os.WriteFile(manifest, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if err := run([]string{"detect", "--manifest", manifest, artifact}, &bytes.Buffer{}); err == nil {
			t.Fatalf("accepted bad manifest %q", text)
		}
	}
}
