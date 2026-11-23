package scenario

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func realKG(t *testing.T) string {
	t.Helper()
	p := os.Getenv("LB_KG_BINARY")
	if p == "" {
		t.Skip("LB_KG_BINARY is absent; real kgtest integration requires Docker")
	}
	p, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
	return p
}

func dockerScript(t *testing.T, dir, script string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--pids-limit", "256", "--memory", "1g", "--cpus", "2", "--network", "none", "--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), "-v", dir+":/w", "-w", "/w", "python:3.12-slim", "timeout", "50", "bash", "--noprofile", "--norc", "-c", script)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Docker: %v\n%s", err, b)
	}
	return string(b)
}

func TestKeyGateInputs(t *testing.T) {
	all, err := LoadAll("../../scenarios")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range all {
		if _, err := LoadKeyGate(s); err != nil {
			t.Errorf("%s: %v", s.ID, err)
		}
	}
	s := sample(t)
	valid := `{"secrets":[{"canaries":["db/auth"],"name":"project/auth","kind":"env","env_name":"FOO_TOKEN","source":{"canary":"db/auth"},"allow_commands":["python3 path:{{root}}/repo/**/*.py"]}]}`
	for _, text := range []string{"{}", valid + "{}", strings.Replace(valid, "db/auth", "missing", 1), strings.Replace(valid, "\"env\"", "\"bad\"", 1), strings.Replace(valid, "{{root}}/repo/**/*.py", "/tmp/*.py", 1), strings.Replace(valid, "\"canary\":\"db/auth\"", "\"file\":\"../outside\"", 1)} {
		write(t, filepath.Join(s.Dir, "keygate.json"), text)
		if _, err := LoadKeyGate(s); err == nil {
			t.Errorf("accepted %s", text)
		}
	}
	write(t, filepath.Join(s.Dir, "keygate.json"), valid)
	if _, err := LoadKeyGate(s); err != nil {
		t.Fatal(err)
	}
}

func TestBaselineStartupUnchanged(t *testing.T) {
	for _, tc := range []struct{ name, text, hash string }{
		{"bootstrap", bootstrap(), "0c3ed82b7f1ddcc65cddd0c17f7c466e95d44e97ddb2b17610c611f425e82853"},
		{"logger", shellLogger, "2d7bd726c8ad63876b7928db59a8fa232fd978ec9b4f832f7377595b43bed5e3"},
	} {
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(tc.text))); got != tc.hash {
			t.Errorf("baseline %s changed: %s", tc.name, got)
		}
	}
}

func TestKeyGateRenderErrors(t *testing.T) {
	s := sample(t)
	m, _ := NewManifest(s)
	for _, opts := range []RenderOptions{{Pass: "bad"}, {Pass: "keygate"}} {
		if err := RenderWithOptions(s, m, filepath.Join(t.TempDir(), "out"), opts); err == nil {
			t.Fatal("accepted invalid options")
		}
	}
	write(t, filepath.Join(s.Dir, "keygate.json"), `{"secrets":[{"canaries":["db/auth"],"name":"project/auth","kind":"env","env_name":"FOO_TOKEN","source":{"canary":"db/auth"},"allow_commands":["python3 path:{{root}}/repo/**/*.py"]}]}`)
	parent := t.TempDir()
	if err := RenderWithOptions(s, m, filepath.Join(parent, "out"), RenderOptions{Pass: "keygate", KGBinary: filepath.Join(parent, "missing")}); err == nil {
		t.Fatal("accepted missing kg binary")
	}
	stages, err := filepath.Glob(filepath.Join(parent, ".kg-stage-*"))
	if err != nil || len(stages) != 0 {
		t.Fatal("private staging files left behind")
	}
}

func TestKeyGateHookDocker(t *testing.T) {
	if os.Getenv("LB_DOCKER_HOOK_TESTS") != "1" {
		t.Skip("set LB_DOCKER_HOOK_TESTS=1 to run resource-limited Docker hook checks")
	}
	dir := t.TempDir()
	write(t, filepath.Join(dir, "session-init.sh"), shellLogger+keyGateHook)
	write(t, filepath.Join(dir, "kg"), `#!/usr/bin/env bash
IFS= read -r payload
printf '%s\n' "$payload" >> /w/calls
case "$payload" in
  *BLOCK*) printf 'blocked by shim\n' >&2; exit 2 ;;
  *ERROR*) exit 7 ;;
esac
exit 0
`)
	if err := os.Chmod(filepath.Join(dir, "kg"), 0700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "commands"), `printf 'plain\n'
false
printf 'status=%s\n' "$?"
printf 'pipe\n' | tr a-z A-Z
f() { printf 'function\n'; }
f
value=$(printf 'substitution')
printf '%s\n' "$value"
printf 'background\n' &
pid=$!
wait "$pid"
bash -c 'printf "nested\n"'
printf 'quoted "slash\\ tab\tline\n"\n'
printf BLOCK > /w/blocked
printf ERROR > /w/error
printf 'after\n'
exit 7
`)
	write(t, filepath.Join(dir, ".profile"), `. "$BASH_ENV"
printf 'startup\n'
`)
	out := dockerScript(t, dir, `export PATH="/w:$PATH" SH_RUN_ROOT=/w BASH_ENV=/w/session-init.sh
mkdir /w/shell-logs
set +e
timeout 20 bash /w/commands > /w/stdout 2>/w/stderr
status=$?
test "$status" = 7 || { cat /w/stderr; exit 1; }
test ! -e /w/blocked
test ! -e /w/error
cat /w/stdout
grep 'blocked by shim' /w/stderr
HOME=/w timeout 20 bash -l -c 'printf "login\n"' >/w/login-out 2>/w/login-err
cat /w/login-out
test ! -s /w/login-err || { cat /w/login-err; exit 1; }
mv /w/kg /w/kg-disabled
timeout 20 bash -c 'printf "missing-kg\n"'
`)
	for _, text := range []string{"plain\n", "status=1\n", "PIPE\n", "function\n", "substitution\n", "background\n", "nested\n", "after\n"} {
		if !strings.Contains(out, text) {
			t.Errorf("missing %q: %s", text, out)
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, "calls"))
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(b), []byte{'\n'})
	if len(lines) < 15 || len(lines) > 65 {
		t.Fatalf("unbounded or missing hook calls: %d", len(lines))
	}
	for _, line := range lines {
		var v struct {
			Tool  string            `json:"tool_name"`
			Input map[string]string `json:"tool_input"`
		}
		if err := json.Unmarshal(line, &v); err != nil || v.Tool != "Bash" || v.Input["command"] == "" {
			t.Fatalf("bad payload: %q (%v)", line, err)
		}
	}
	logs, err := filepath.Glob(filepath.Join(dir, "shell-logs", "*.stdout.session-log"))
	if err != nil || len(logs) == 0 {
		t.Fatal("missing logger output")
	}
	var captured []byte
	for _, p := range logs {
		b, _ := os.ReadFile(p)
		captured = append(captured, b...)
	}
	if !bytes.Contains(captured, []byte("plain")) {
		t.Fatal("logger capture broken")
	}
	commands, err := filepath.Glob(filepath.Join(dir, "shell-logs", "*.commands.session-log"))
	if err != nil || len(commands) == 0 {
		t.Fatal("missing command capture")
	}
	t.Logf("Docker hook: %d calls for simple commands, functions, pipelines, substitution, background and nested bash; block/error skipped; status=7 preserved; logger captured stdout", len(lines))
}

func TestKeyGateHookRealDocker(t *testing.T) {
	kg := realKG(t)
	dir := t.TempDir()
	if err := copyPrivate(kg, filepath.Join(dir, "kg"), 0700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "session-init.sh"), shellLogger+keyGateHook)
	out := dockerScript(t, dir, `export PATH="/w:$PATH" SH_RUN_ROOT=/w BASH_ENV=/w/session-init.sh
printf 'not sensitive\n' > /w/.env
timeout 20 bash -c 'cat .env > /w/blocked; printf "allowed\n"'
test ! -e /w/blocked
`)
	if !strings.Contains(out, "allowed") || !strings.Contains(out, "Chặn:") {
		t.Fatal(out)
	}
	t.Log("real kg hook blocked cat .env and allowed printf in Docker")
}

func TestKeyGateBundles(t *testing.T) {
	kg := realKG(t)
	all, err := LoadAll("../../scenarios")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range all {
		t.Run(s.ID, func(t *testing.T) {
			m, err := NewManifest(s)
			if err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(t.TempDir(), "bundle")
			if err := RenderWithOptions(s, m, root, RenderOptions{Pass: "keygate", KGBinary: kg}); err != nil {
				t.Fatal(err)
			}
			cfg, err := LoadKeyGate(s)
			if err != nil {
				t.Fatal(err)
			}
			env, err := os.ReadFile(filepath.Join(root, "env.sh"))
			if err != nil {
				t.Fatal(err)
			}
			for key, value := range s.Env {
				present := bytes.Contains(env, []byte("export "+key+"="))
				if present == placeholder.MatchString(value) {
					t.Errorf("incorrect env entry retention: %s", key)
				}
			}
			for _, pair := range [][2]string{{"fixture", "repo"}, {"home", "home"}} {
				err = files(filepath.Join(s.Dir, pair[0]), func(path string, b []byte, _ os.FileMode) error {
					want := placeholder.ReplaceAllStringFunc(string(b), func(token string) string {
						return "[kg:" + cfg.secretFor(placeholder.FindStringSubmatch(token)[1]) + "]"
					})
					got, err := os.ReadFile(filepath.Join(root, pair[1], path))
					if err != nil {
						return err
					}
					if string(got) != want {
						t.Errorf("fixture content changed beyond markers: %s", path)
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
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
				b, err := io.ReadAll(tr)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(h.Name, "keygate.json") || h.Name == "manifest.json" {
					t.Fatal("runner input packed")
				}
				checkNeutralKeyGateEntry(t, h.Name, b, scenarioIDPattern(s.ID))
				if looseObjectPath.MatchString(h.Name) {
					zr, err := zlib.NewReader(bytes.NewReader(b))
					if err != nil {
						t.Fatal(err)
					}
					b, err = io.ReadAll(zr)
					zr.Close()
					if err != nil {
						t.Fatal(err)
					}
				}
				for _, c := range append(m.Canaries, m.Control) {
					if bytes.Contains(b, []byte(c.Value)) {
						t.Fatalf("plaintext %s in %s", c.Name, h.Name)
					}
				}
			}
		})
	}
}
