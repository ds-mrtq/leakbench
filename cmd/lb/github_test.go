package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ds-mrtq/leakbench/internal/ghguard"
)

func guardFixture() ghguard.Snapshot {
	stamp := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	return ghguard.Snapshot{Version: 1, StartedAt: stamp, FinishedAt: stamp, Scope: "explicit", Repos: map[string]ghguard.Repository{
		"owner/project": {Refs: map[string]string{}, PullRequests: map[string]json.RawMessage{}, Issues: map[string]json.RawMessage{}, Releases: map[string]json.RawMessage{}},
	}}
}

func TestGitHubDiffCommand(t *testing.T) {
	dir := t.TempDir()
	before := filepath.Join(dir, "before.json")
	after := filepath.Join(dir, "after.json")
	if err := ghguard.Save(before, guardFixture()); err != nil {
		t.Fatal(err)
	}
	if err := ghguard.Save(after, guardFixture()); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run([]string{"ghdiff", before, after}, &out); err != nil || out.String() != "no GitHub side effects\n" {
		t.Fatal(out.String(), err)
	}
	if err := run([]string{"ghdiff", before, after}, brokenWriter{}); err == nil {
		t.Fatal("missing output error")
	}
	client := ghguard.Client{Run: func(name string, args ...string) ([]byte, error) {
		if name != "gh" || !strings.Contains(args[len(args)-1], "/events?") {
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		return []byte(`[{"id":"123","type":"DeleteEvent","actor":{"login":"operator"},"created_at":"2026-01-02T00:00:01Z","payload":{"ref":"temporary"}}]`), nil
	}}
	out.Reset()
	if err := runGitHub([]string{"ghdiff", "--since", "2026-01-02T00:00:00Z", before, after}, &out, client); !errors.Is(err, errGitHubChanged) || !strings.Contains(out.String(), `actor="operator"`) || !strings.Contains(out.String(), `ref="temporary"`) {
		t.Fatal(out.String(), err)
	}
	client.Run = func(string, ...string) ([]byte, error) { return nil, errors.New("denied") }
	out.Reset()
	if err := runGitHub([]string{"ghdiff", "--since", "2026-01-02T00:00:00Z", before, after}, &out, client); err == nil || strings.Contains(out.String(), "no GitHub side effects") {
		t.Fatal("event failure reported as clean")
	}
	changed := guardFixture()
	changed.Repos["owner/project"].Issues["1"] = json.RawMessage(`{"number":1,"comments":1}`)
	if err := os.Remove(after); err != nil {
		t.Fatal(err)
	}
	if err := ghguard.Save(after, changed); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := run([]string{"ghdiff", before, after}, &out); !errors.Is(err, errGitHubChanged) || !strings.Contains(out.String(), "issue #1 added") {
		t.Fatal(out.String(), err)
	}
	if err := run([]string{"ghdiff", before, after}, brokenWriter{}); err == nil {
		t.Fatal("missing diff output error")
	}
	for _, args := range [][]string{
		{"ghdiff"}, {"ghdiff", "--bad"}, {"ghdiff", "--since", "bad", before, after}, {"ghdiff", "missing", after}, {"ghdiff", before, "missing"}, {"ghsnap"}, {"ghsnap", "--bad"},
	} {
		if err := run(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestGitHubSnapshotCommand(t *testing.T) {
	client := ghguard.Client{Run: func(name string, args ...string) ([]byte, error) {
		if name == "git" {
			return nil, nil
		}
		if name != "gh" || !strings.Contains(strings.Join(args, " "), "--method GET") {
			t.Fatal("not a read-only API command")
		}
		return []byte("[]"), nil
	}}
	dest := filepath.Join(t.TempDir(), "snapshot.json")
	var out bytes.Buffer
	if err := runGitHub([]string{"ghsnap", "--out", dest, "--repos", "owner/project"}, &out, client); err != nil || !strings.Contains(out.String(), "1 repositories (explicit)") {
		t.Fatal(out.String(), err)
	}
	if _, err := ghguard.Load(dest); err != nil {
		t.Fatal(err)
	}
	if err := runGitHub([]string{"ghsnap", "--out", dest, "--repos", "owner/project"}, &out, client); err == nil {
		t.Fatal("overwrote existing snapshot")
	}
	if err := runGitHub([]string{"ghsnap", "--out", dest + "new", "--repos", "owner/project"}, brokenWriter{}, client); err == nil {
		t.Fatal("missing snapshot output error")
	}
	if err := runGitHub([]string{"ghsnap", "--out", dest, "--repos", "../bad"}, &out, client); err == nil {
		t.Fatal("accepted unsafe repository")
	}
	client.Run = func(string, ...string) ([]byte, error) { return nil, errors.New("failed") }
	if err := runGitHub([]string{"ghsnap", "--out", dest + "failed"}, &out, client); err == nil {
		t.Fatal("ignored snapshot command failure")
	}
	if _, err := os.Stat(dest + "failed"); !os.IsNotExist(err) {
		t.Fatal("created incomplete snapshot")
	}
}
