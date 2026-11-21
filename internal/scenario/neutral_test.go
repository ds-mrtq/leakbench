package scenario

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var identifyingText = regexp.MustCompile(`(?i)leakbench|canar(?:y|ies)|benchmark|shell-logger|(^|[^a-z0-9])lb($|[^a-z0-9])`)
var looseObjectPath = regexp.MustCompile(`/objects/[0-9a-f]{2}/[0-9a-f]{38}$`)

func TestIdentifyingText(t *testing.T) {
	for _, text := range []string{"LeakBench", "CANARY", "canaries", "Benchmark", "shell-logger.sh", ".lb-log", "_LB_CAPTURE_PID", "LB_RUN_ROOT", "lb", "LB-SHELL-1"} {
		if !identifyingText.MatchString(text) {
			t.Errorf("missed %q", text)
		}
	}
	for _, text := range []string{"club", "albatross", "shell-logs", "SH_RUN_ROOT", "_SH_TRACE_PID", "SH-SESSION-1", ".session-log"} {
		if identifyingText.MatchString(text) {
			t.Errorf("false positive: %q", text)
		}
	}
}

func scenarioIDPattern(id string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)(^|[^a-z0-9])` + regexp.QuoteMeta(id) + `($|[^a-z0-9])`)
}

func TestScenarioIDPattern(t *testing.T) {
	pattern := scenarioIDPattern("A01")
	for _, text := range []string{"A01", "remote/a01.git", "_a01_capture", "task A01.", "100644 A01.txt", "message\nA01\n"} {
		if !pattern.MatchString(text) {
			t.Errorf("missed scenario ID in %q", text)
		}
	}
	for _, text := range []string{"", "A02", "remote/project.git", "XA01", "A01x", "01fd352ba01e6f9df89bc66b5a61eb4a8b79b5"} {
		if pattern.MatchString(text) {
			t.Errorf("false positive in %q", text)
		}
	}
}

func checkNeutralContent(t *testing.T, name string, b []byte, id *regexp.Regexp) {
	t.Helper()
	if match := identifyingText.Find(b); match != nil {
		t.Errorf("identifying text %q in %s", match, name)
	}
	if match := id.Find(b); match != nil {
		t.Errorf("scenario ID %q in %s", match, name)
	}
}

func checkNeutralGitIndex(t *testing.T, name string, b []byte, id *regexp.Regexp) {
	t.Helper()
	if len(b) < 12 || string(b[:4]) != "DIRC" || binary.BigEndian.Uint32(b[4:8]) != 2 {
		t.Fatal("unexpected Git index format")
	}
	count := binary.BigEndian.Uint32(b[8:12])
	data := b[12:]
	for i := uint32(0); i < count; i++ {
		if len(data) < 62 {
			t.Fatal("truncated Git index entry")
		}
		end := bytes.IndexByte(data[62:], 0)
		if end < 0 {
			t.Fatal("missing Git index path")
		}
		checkNeutralContent(t, name+" (index path)", data[62:62+end], id)
		size := (62 + end + 1 + 7) / 8 * 8
		if len(data) < size {
			t.Fatal("truncated Git index padding")
		}
		data = data[size:]
	}
}

func checkNeutralGitObject(t *testing.T, name string, b []byte, id *regexp.Regexp) {
	t.Helper()
	zr, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	data, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	header, data, ok := bytes.Cut(data, []byte{0})
	if !ok {
		t.Fatal("invalid Git object")
	}
	if !bytes.HasPrefix(header, []byte("tree ")) {
		checkNeutralContent(t, name+" (inflated)", data, id)
		return
	}
	// Tree entries contain binary object IDs, which are not visible text.
	for len(data) > 0 {
		entry, rest, ok := bytes.Cut(data, []byte{0})
		if !ok || len(rest) < 20 {
			t.Fatal("invalid Git tree")
		}
		checkNeutralContent(t, name+" (tree entry)", entry, id)
		data = rest[20:]
	}
}

func TestNeutralGitMetadataHashes(t *testing.T) {
	id := scenarioIDPattern("A01")
	index := make([]byte, 12+72+20)
	copy(index, "DIRC")
	binary.BigEndian.PutUint32(index[4:8], 2)
	binary.BigEndian.PutUint32(index[8:12], 1)
	copy(index[12+40:], "A01")
	copy(index[12+62:], "README.md")
	copy(index[len(index)-20:], "A01")
	checkNeutralGitIndex(t, "index", index, id)

	body := append([]byte("100644 README.md\x00"), make([]byte, 20)...)
	copy(body[len(body)-20:], "A01")
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := fmt.Fprintf(zw, "tree %d\x00%s", len(body), body); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	checkNeutralGitObject(t, "tree", compressed.Bytes(), id)
}

func TestRenderedBundlesAreNeutral(t *testing.T) {
	all, err := LoadAll(filepath.Join("..", "..", "scenarios"))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 46 {
		t.Fatalf("want 46 scenarios, got %d", len(all))
	}
	for _, s := range all {
		t.Run(s.ID, func(t *testing.T) {
			id := scenarioIDPattern(s.ID)
			m, err := NewManifest(s)
			if err != nil {
				t.Fatal(err)
			}
			for i := range m.Canaries {
				m.Canaries[i].Value = fmt.Sprintf("%032d", i+1)
			}
			m.Control.Value = strings.Repeat("9", 32)
			root := filepath.Join(t.TempDir(), "project")
			if err := Render(s, m, root); err != nil {
				t.Fatal(err)
			}
			f, err := os.Open(filepath.Join(root, "bundle.tar.gz"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			gz, err := gzip.NewReader(f)
			if err != nil {
				t.Fatal(err)
			}
			defer gz.Close()
			tr := tar.NewReader(gz)
			for {
				h, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				if h.Name == "manifest.json" || h.Name == "collection-key.pem" {
					t.Fatal("runner-only file in bundle")
				}
				checkNeutralContent(t, h.Name+" (filename)", []byte(h.Name), id)
				data, err := io.ReadAll(tr)
				if err != nil {
					t.Fatal(err)
				}
				if looseObjectPath.MatchString(h.Name) {
					checkNeutralGitObject(t, h.Name, data, id)
				} else if strings.HasSuffix(h.Name, "/.git/index") {
					checkNeutralGitIndex(t, h.Name, data, id)
				} else {
					checkNeutralContent(t, h.Name, data, id)
				}
			}
		})
	}
}
