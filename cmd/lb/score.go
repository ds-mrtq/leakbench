package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ds-mrtq/leakbench/internal/scenario"
	"github.com/ds-mrtq/leakbench/internal/score"
)

func runScore(args []string, out io.Writer) error {
	f := flag.NewFlagSet("score", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	root := f.String("runs", "", "archived run directories")
	dest := f.String("out", "", "score output directory")
	scenarios := f.String("scenarios", "scenarios", "scenario root")
	if err := f.Parse(args); err != nil {
		return err
	}
	if *root == "" || *dest == "" || f.NArg() != 0 {
		return fmt.Errorf("usage: lb score --runs DIR --out DIR [--scenarios scenarios]")
	}
	entries, err := os.ReadDir(*root)
	if err != nil {
		return err
	}
	var runs []score.Result
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		s, err := scenario.Load(filepath.Join(*scenarios, e.Name()))
		if err != nil {
			return err
		}
		r, err := score.Run(filepath.Join(*root, e.Name()), s.Class)
		if err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
		runs = append(runs, r)
	}
	if len(runs) == 0 {
		return fmt.Errorf("no runs found")
	}
	summary := score.Summarize(runs)
	if err := score.Write(*dest, summary); err != nil {
		return err
	}
	_, err = io.WriteString(out, score.Markdown(summary))
	return err
}
