package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/ds-mrtq/leakbench/internal/scenario"
)

func runRender(args []string, out io.Writer) error {
	f := flag.NewFlagSet("render", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	dir := f.String("scenario", "", "scenario directory")
	dest := f.String("out", "", "empty output directory")
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
	if err = scenario.Render(s, m, *dest); err != nil {
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
	_, err = fmt.Fprintf(out, "Linted %d scenarios\n", len(all))
	return err
}
