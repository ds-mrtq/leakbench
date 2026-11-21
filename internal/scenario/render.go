package scenario

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ds-mrtq/leakbench/internal/canary"
	"github.com/ds-mrtq/leakbench/internal/sealed"
)

// Render writes an inspectable bundle tree and bundle.tar.gz. The manifest is
// deliberately outside the archive: it is runner state, not agent input.
// outDir must be empty to avoid overwriting a previous run.
func Render(s Scenario, m canary.Manifest, outDir string) error {
	if err := Lint(s); err != nil {
		return err
	}
	values := map[string]string{}
	for _, c := range m.Canaries {
		if _, ok := values[c.Name]; ok || len(c.Value) < 16 || c.Value == m.Control.Value {
			return fmt.Errorf("invalid manifest canary %q", c.Name)
		}
		values[c.Name] = c.Value
	}
	if len(values) != len(s.Canaries) || len(m.Control.Value) < 16 || m.Control.Name != "control" {
		return fmt.Errorf("manifest does not match scenario")
	}
	for _, d := range s.Canaries {
		found := false
		for _, c := range m.Canaries {
			if c.Name == d.Name && c.Kind == d.Kind {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("manifest missing %s (%s)", d.Name, d.Kind)
		}
	}
	entries, err := os.ReadDir(outDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("output directory must be empty")
	}
	if err = os.MkdirAll(outDir, 0700); err != nil {
		return err
	}
	root, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}
	privateKey, publicKey, err := sealed.NewKeyPair()
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(root, "collection-key.pem"), privateKey, 0600); err != nil {
		return err
	}
	replace := func(text string) string {
		return placeholder.ReplaceAllStringFunc(text, func(token string) string { return values[placeholder.FindStringSubmatch(token)[1]] })
	}
	for _, pair := range [][2]string{{"fixture", "repo"}, {"home", "home"}} {
		dest := filepath.Join(root, pair[1])
		if err = os.MkdirAll(dest, 0700); err != nil {
			return err
		}
		err = files(filepath.Join(s.Dir, pair[0]), func(path string, b []byte, mode fs.FileMode) error {
			target := filepath.Join(dest, path)
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				return err
			}
			return os.WriteFile(target, []byte(replace(string(b))), 0600|mode&0111)
		})
		if err != nil {
			return err
		}
	}
	var env strings.Builder
	keys := make([]string, 0, len(s.Env))
	for key := range s.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(&env, "export %s=%s\n", key, shellQuote(replace(s.Env[key])))
	}
	for name, content := range map[string]string{"env.sh": env.String(), "bootstrap.sh": bootstrap(), "collect.sh": collect(publicKey), "session-init.sh": shellLogger} {
		mode := fs.FileMode(0600)
		if strings.HasSuffix(name, ".sh") {
			mode = 0700
		}
		if err = os.WriteFile(filepath.Join(root, name), []byte(content), mode); err != nil {
			return err
		}
	}
	repo := filepath.Join(root, "repo")
	if err = git(repo, "init", "-b", "main"); err != nil {
		return err
	}
	for _, args := range [][]string{{"config", "user.name", "Local Developer"}, {"config", "user.email", "developer@example.test"}, {"config", "core.hooksPath", "/dev/null"}, {"config", "commit.gpgsign", "false"}} {
		if err = git(repo, args...); err != nil {
			return err
		}
	}
	// Local configuration and logs are present but not part of the baseline commit.
	if err = files(filepath.Join(s.Dir, "fixture"), func(path string, b []byte, _ fs.FileMode) error {
		if placeholder.Match(b) {
			return nil
		}
		return git(repo, "add", "--", path)
	}); err != nil {
		return err
	}
	if err = git(repo, "commit", "--allow-empty", "-m", "Initial project"); err != nil {
		return err
	}
	remote := filepath.Join(root, "remote", "project.git")
	if err = os.MkdirAll(filepath.Dir(remote), 0700); err != nil {
		return err
	}
	if err = git(root, "clone", "--bare", "--no-hardlinks", repo, remote); err != nil {
		return err
	}
	if err = git(remote, "remote", "remove", "origin"); err != nil {
		return err
	}
	if err = git(repo, "remote", "add", "origin", "~/remote/project.git"); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(root, "manifest.json"), append(b, '\n'), 0600); err != nil {
		return err
	}
	return pack(root)
}

func git(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-c", "init.templateDir=", "-c", "core.hooksPath=/dev/null"}, args...)...)
	cmd.Dir = dir
	// Do not inherit caller Git redirections, identity, signing or config.
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GIT_") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s: %w: %s", args[0], err, output)
	}
	return nil
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }

func bootstrap() string {
	return `#!/usr/bin/env bash
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
: "${HOME:?HOME must point to the isolated run home}"
case "$HOME" in /*) ;; *) echo 'HOME must be absolute' >&2; exit 1 ;; esac
test "$HOME" != / || exit 1
mkdir -p "$HOME" "$HOME/remote" "$HOME/shell-logs" "$root/shell-logs"
remote="$HOME/remote/project.git"
if [ -e "$remote" ]; then
  echo 'Remote already exists; use a fresh run home' >&2
  exit 1
fi
cp -R "$root/home/." "$HOME/"
printf 'export SH_RUN_ROOT=%q\nexport BASH_ENV=%q\n. "$BASH_ENV"\n' "$root" "$root/session-init.sh" > "$HOME/.session-init"
for profile in .bashrc .profile .bash_profile .bash_login; do
  case "$profile" in .bash_profile|.bash_login) [ -f "$HOME/$profile" ] || continue ;; esac
  printf '\n' >> "$HOME/$profile"
  cat "$root/env.sh" >> "$HOME/$profile"
  printf '\nif [ -n "${BASH_VERSION-}" ]; then . "$HOME/.session-init"; fi\n' >> "$HOME/$profile"
done
cp -R "$root/remote/project.git" "$remote"
cd "$root/repo"
git init -q -b main
git config --local user.name 'Local Developer'
git config --local user.email 'developer@example.test'
git config --local commit.gpgsign false
git config --local core.hooksPath /dev/null
if ! git rev-parse --verify HEAD >/dev/null 2>&1; then
  git add --all
  git commit -qm 'Initial project'
fi
git remote set-url origin "$remote"
. "$root/env.sh"
printf 'Workspace: %s/repo\nFor subsequent commands: . "$HOME/.profile"\n' "$root"
`
}

func collect(publicKey []byte) string {
	var encoded strings.Builder
	for _, b := range publicKey {
		fmt.Fprintf(&encoded, `\x%02x`, b)
	}
	return `#!/usr/bin/env bash
set +x
set -euo pipefail
umask 077
if ! command -v openssl >/dev/null 2>&1; then
  printf '%s\n' 'openssl is required to create the session archive' >&2
  exit 127
fi
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
: "${HOME:?HOME must point to the isolated run home}"
if [ -n "${EPOCHREALTIME:-}" ]; then
  { (set -C; printf 'SH-COLLECTION-1\t%s\n' "$EPOCHREALTIME" > "$root/shell-logs/collection-start"); } 2>/dev/null || :
fi
out="$root/session-archive.bin"
stage=$(mktemp -d "$root/.collect.XXXXXX")
trap 'rm -rf "$stage"' EXIT
trap 'exit 1' HUP INT TERM
mkdir "$stage/home" "$stage/run"
for path in "$root/shell-logs" "$root/"*.log "$root/typescript"; do
  if [ -e "$path" ]; then cp -R "$path" "$stage/run/"; fi
done
for path in "$HOME/.bash_history" "$HOME/.zsh_history" "$HOME/.bash_sessions" "$HOME/shell-logs" "$HOME/"*.log "$HOME/typescript"; do
  if [ -e "$path" ]; then cp -R "$path" "$stage/home/"; fi
done
tar -czf "$stage/archive.tar.gz" -C "$root" repo -C "$HOME" remote/project.git -C "$stage" home run
printf '%b' '` + encoded.String() + `' > "$stage/recipient.der"
openssl rand -out "$stage/material" 64
openssl rand -out "$stage/iv" 16
material=$(od -An -v -tx1 "$stage/material" | tr -d ' \n')
iv=$(od -An -v -tx1 "$stage/iv" | tr -d ' \n')
openssl pkeyutl -encrypt -pubin -keyform DER -inkey "$stage/recipient.der" -in "$stage/material" -out "$stage/wrapped" -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:sha256 -pkeyopt rsa_mgf1_md:sha256
openssl enc -aes-256-cbc -K "${material:0:64}" -iv "$iv" -in "$stage/archive.tar.gz" -out "$stage/body"
{ printf '` + sealed.Magic + `'; cat "$stage/wrapped" "$stage/iv" "$stage/body"; } > "$stage/payload"
openssl dgst -sha256 -mac HMAC -macopt "hexkey:${material:64:64}" -binary "$stage/payload" > "$stage/tag"
cat "$stage/payload" "$stage/tag" > "$stage/sealed"
mv "$stage/sealed" "$out"
rm -rf "$stage"
printf '%s\n' "$out"
`
}

func pack(root string) (err error) {
	f, err := os.OpenFile(filepath.Join(root, "bundle.tar.gz"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	defer func() {
		for _, close := range []func() error{tw.Close, gz.Close, f.Close} {
			if e := close(); err == nil {
				err = e
			}
		}
	}()
	for _, name := range []string{"repo", "home", "remote", "env.sh", "bootstrap.sh", "collect.sh", "session-init.sh"} {
		err = filepath.WalkDir(filepath.Join(root, name), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			h, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			h.Name = filepath.ToSlash(rel)
			if err = tw.WriteHeader(h); err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			in, err := os.Open(path)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(tw, in)
			closeErr := in.Close()
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		})
		if err != nil {
			return err
		}
	}
	return nil
}
