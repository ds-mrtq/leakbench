package scenario

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderKeysFresh(t *testing.T) {
	s := sample(t)
	m, err := NewManifest(s)
	if err != nil {
		t.Fatal(err)
	}
	var previous []byte
	for i := 0; i < 2; i++ {
		root := filepath.Join(t.TempDir(), "run")
		if err := Render(s, m, root); err != nil {
			t.Fatal(err)
		}
		key, err := os.ReadFile(filepath.Join(root, "collection-key.pem"))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(key, previous) {
			t.Fatal("render reused private key")
		}
		previous = key
	}
}

func TestCollectionRequiresOpenSSL(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	write(t, filepath.Join(root, "collect.sh"), collect(nil))
	cmd := exec.Command(bash, filepath.Join(root, "collect.sh"))
	cmd.Env = []string{"PATH=" + t.TempDir(), "HOME=" + root}
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "openssl is required") {
		t.Fatalf("missing clear openssl error: %v %s", err, out)
	}
}
