package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScenarioCommands(t *testing.T) {
	root := filepath.Join("..", "..", "scenarios")
	var out bytes.Buffer
	if err := run([]string{"lint", "--scenarios", root}, &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Linted 10 scenarios\n" {
		t.Fatalf("unexpected lint output %q", out.String())
	}
	dest := filepath.Join(t.TempDir(), "rendered")
	out.Reset()
	if err := run([]string{"render", "--scenario", filepath.Join(root, "A01"), "--out", dest}, &out); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"manifest.json", "bundle.tar.gz", "bootstrap.sh", "collect.sh"} {
		if _, err := os.Stat(filepath.Join(dest, name)); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(out.String(), "Rendered A01") {
		t.Fatal("missing render result")
	}
	for _, args := range [][]string{
		{"render"}, {"render", "--bad"}, {"render", "--scenario", "missing", "--out", dest},
		{"render", "--scenario", filepath.Join(root, "A01"), "--out", dest},
		{"lint", "extra"}, {"lint", "--bad"}, {"lint", "--scenarios", t.TempDir()},
	} {
		if err := run(args, &out); err == nil {
			t.Fatalf("accepted invalid arguments %v", args)
		}
	}
	if err := run([]string{"lint", "--scenarios", root}, brokenWriter{}); err == nil {
		t.Fatal("missing output error")
	}
}
