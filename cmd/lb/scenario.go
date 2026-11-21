package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ds-mrtq/leakbench/internal/scenario"
)

func runRender(args []string, out io.Writer) error {
	f := flag.NewFlagSet("render", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	dir := f.String("scenario", "", "scenario directory")
	dest := f.String("out", "", "empty output directory")
	pass := f.String("pass", "baseline", "baseline|keygate")
	kg := f.String("kg-binary", os.Getenv("LB_KG_BINARY"), "pinned kgtest binary (or LB_KG_BINARY)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if *dir == "" || *dest == "" || f.NArg() != 0 {
		return fmt.Errorf("usage: lb render --scenario scenarios/A01 --out DIR")
	}
	s, err := scenario.Load(*dir)
	if err != nil {
		return err
	}
	m, err := scenario.NewManifest(s)
	if err != nil {
		return err
	}
	if err = scenario.RenderWithOptions(s, m, *dest, scenario.RenderOptions{Pass: *pass, KGBinary: *kg}); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Rendered %s to %s (manifest.json and bundle.tar.gz)\n", s.ID, *dest)
	return err
}

func runLint(args []string, out io.Writer) error {
	f := flag.NewFlagSet("lint", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	root := f.String("scenarios", "scenarios", "scenario root")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("usage: lb lint [--scenarios DIR]")
	}
	all, err := scenario.LoadAll(*root)
	if err != nil {
		return err
	}
	for _, s := range all {
		if _, err := scenario.LoadKeyGate(s); err != nil {
			return fmt.Errorf("%s: %w", s.ID, err)
		}
	}
	_, err = fmt.Fprintf(out, "Linted %d scenarios\n", len(all))
	return err
}
