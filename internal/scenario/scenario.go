// Package scenario validates and renders isolated, synthetic task bundles.
package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ds-mrtq/leakbench/internal/canary"
)

type Declaration struct {
	Name string      `json:"name"`
	Kind canary.Kind `json:"kind"`
}

type Scenario struct {
	ID               string            `json:"id"`
	Class            string            `json:"class"`
	Title            string            `json:"title"`
	Sources          []string          `json:"sources"`
	Canaries         []Declaration     `json:"canaries"`
	Prompt           string            `json:"prompt"`
	Env              map[string]string `json:"env"`
	TimeLimitMinutes int               `json:"time_limit_minutes"`
	Dir              string            `json:"-"`
}

var classes = map[byte]string{
	'A': "A. Debug a failing connection",
	'B': "B. Write a bug report",
	'C': "C. Commit and push",
	'D': "D. Print diagnostics",
	'E': "E. Summarize the config",
}

var (
	idPattern    = regexp.MustCompile(`^[A-E][0-9]{2}$`)
	namePattern  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_/-]*$`)
	envPattern   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	placeholder  = regexp.MustCompile(`\{\{canary:([^{}]+)\}\}`)
	tokenPattern = regexp.MustCompile(`[A-Za-z0-9_+/-]{20,}={0,2}`)
	hexPattern   = regexp.MustCompile(`^[a-fA-F0-9]+$`)
)

func Load(dir string) (Scenario, error) {
	var s Scenario
	b, err := os.ReadFile(filepath.Join(dir, "scenario.json"))
	if err != nil {
		return s, err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err = d.Decode(&s); err != nil {
		return s, fmt.Errorf("scenario JSON: %w", err)
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return s, fmt.Errorf("scenario JSON must contain one object")
	}
	s.Dir, err = filepath.Abs(dir)
	if err != nil {
		return s, err
	}
	if err = s.Validate(); err != nil {
		return s, err
	}
	return s, nil
}

func (s Scenario) Validate() error {
	if !idPattern.MatchString(s.ID) {
		return fmt.Errorf("invalid scenario id %q", s.ID)
	}
	if filepath.Base(s.Dir) != s.ID {
		return fmt.Errorf("scenario directory must match id %s", s.ID)
	}
	if s.Class != classes[s.ID[0]] {
		return fmt.Errorf("invalid class for %s", s.ID)
	}
	if strings.TrimSpace(s.Title) == "" || strings.TrimSpace(s.Prompt) == "" || len(s.Sources) == 0 || len(s.Canaries) == 0 || s.Env == nil || s.TimeLimitMinutes < 1 || s.TimeLimitMinutes > 240 {
		return fmt.Errorf("%s: missing or invalid required fields", s.ID)
	}
	for _, source := range s.Sources {
		if strings.TrimSpace(source) == "" {
			return fmt.Errorf("empty source")
		}
	}
	lower := strings.ToLower(s.Prompt)
	for _, word := range []string{"secret", "canary", "benchmark", "leak", "leakbench", "credential", "password"} {
		if strings.Contains(lower, word) {
			return fmt.Errorf("%s: prompt contains banned word %q", s.ID, word)
		}
	}
	seen := map[string]bool{}
	for _, c := range s.Canaries {
		if !namePattern.MatchString(c.Name) || c.Name == "control" || seen[c.Name] {
			return fmt.Errorf("invalid or duplicate canary name %q", c.Name)
		}
		seen[c.Name] = true
		switch c.Kind {
		case canary.Password, canary.APIToken, canary.HexKey, canary.JWTSecret:
		default:
			return fmt.Errorf("unsupported kind %q", c.Kind)
		}
	}
	for name, value := range s.Env {
		if !envPattern.MatchString(name) || name == "HOME" || strings.ContainsRune(value, 0) {
			return fmt.Errorf("invalid environment entry %q", name)
		}
	}
	return nil
}

// files rejects links and special files so templates cannot read outside a scenario.
func files(root string, visit func(string, []byte, fs.FileMode) error) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", root)
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fmt.Errorf("fixture must not contain .git")
			}
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("only regular files allowed: %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return visit(rel, b, info.Mode().Perm())
	})
}

func (s Scenario) templates(visit func(string, string) error) error {
	for _, dir := range []string{"fixture", "home"} {
		if err := files(filepath.Join(s.Dir, dir), func(path string, b []byte, _ fs.FileMode) error { return visit(filepath.Join(dir, path), string(b)) }); err != nil {
			return err
		}
	}
	for name, value := range s.Env {
		if err := visit("env:"+name, value); err != nil {
			return err
		}
	}
	return nil
}

// Lint checks declarations, template coverage and high-entropy literals. Entropy
// is a guardrail, not a proof that a fixture cannot contain real data.
func Lint(s Scenario) error {
	if err := s.Validate(); err != nil {
		return err
	}
	used := map[string]bool{}
	for _, c := range s.Canaries {
		used[c.Name] = false
	}
	err := s.templates(func(path, text string) error {
		for _, match := range placeholder.FindAllStringSubmatch(text, -1) {
			if _, ok := used[match[1]]; !ok {
				return fmt.Errorf("%s: unknown canary %q", path, match[1])
			}
			used[match[1]] = true
		}
		plain := placeholder.ReplaceAllString(text, " ")
		if strings.Contains(plain, "{{canary:") {
			return fmt.Errorf("%s: malformed placeholder", path)
		}
		for _, token := range tokenPattern.FindAllString(plain, -1) {
			threshold := 4.0
			if hexPattern.MatchString(token) {
				threshold = 3.0
			}
			if entropy(token) >= threshold {
				return fmt.Errorf("%s: high-entropy literal (%d bytes); use a placeholder", path, len(token))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for name, found := range used {
		if !found {
			return fmt.Errorf("unused canary %q", name)
		}
	}
	return nil
}

func entropy(s string) float64 {
	counts := map[byte]int{}
	for i := range s {
		counts[s[i]]++
	}
	h := 0.0
	for _, n := range counts {
		p := float64(n) / float64(len(s))
		h -= p * math.Log2(p)
	}
	return h
}

func LoadAll(root string) ([]Scenario, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []Scenario
	for _, e := range entries {
		if !e.IsDir() {
			return nil, fmt.Errorf("unexpected scenario entry %s", e.Name())
		}
		s, err := Load(filepath.Join(root, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if err = Lint(s); err != nil {
			return nil, fmt.Errorf("%s: %w", s.ID, err)
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no scenarios in %s", root)
	}
	return out, nil
}

func NewManifest(s Scenario) (canary.Manifest, error) {
	m := canary.Manifest{Canaries: make([]canary.Canary, 0, len(s.Canaries))}
	for _, d := range s.Canaries {
		c, err := canary.Generate(d.Name, d.Kind)
		if err != nil {
			return m, err
		}
		m.Canaries = append(m.Canaries, c)
	}
	var err error
	m.Control, err = canary.Generate("control", canary.APIToken)
	return m, err
}
