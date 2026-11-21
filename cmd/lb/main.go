package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ds-mrtq/leakbench/internal/canary"
	"github.com/ds-mrtq/leakbench/internal/detect"
	"github.com/ds-mrtq/leakbench/internal/ghguard"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: lb <detect|render|lint|score|ghsnap|ghdiff> [options]")
	}
	switch args[0] {
	case "detect":
		return runDetect(args[1:], out)
	case "render":
		return runRender(args[1:], out)
	case "lint":
		return runLint(args[1:], out)
	case "score":
		return runScore(args[1:], out)
	case "ghsnap", "ghdiff":
		return runGitHub(args, out, ghguard.Client{})
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func runDetect(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("detect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("manifest", "", "canary manifest JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *path == "" || flags.NArg() == 0 {
		return errors.New("detect requires --manifest and at least one file")
	}
	b, err := os.ReadFile(*path)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	var m canary.Manifest
	if err = json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	if len(m.Canaries) == 0 {
		return errors.New("manifest has no canaries")
	}
	seen := map[string]bool{}
	for _, c := range append(append([]canary.Canary{}, m.Canaries...), m.Control) {
		if c.Name == "" || len(c.Value) < 16 || seen[c.Name] {
			return errors.New("manifest requires unique canary names and values of at least 16 bytes, including control")
		}
		seen[c.Name] = true
	}
	artifacts := make([]detect.Artifact, 0, flags.NArg())
	for _, path := range flags.Args() {
		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read artifact %s: %w", path, err)
		}
		artifacts = append(artifacts, detect.Artifact{Name: "file:" + path, Data: b})
	}
	return json.NewEncoder(out).Encode(detect.Scan(m, artifacts))
}
