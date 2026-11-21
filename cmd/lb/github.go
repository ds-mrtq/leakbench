package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ds-mrtq/leakbench/internal/ghguard"
)

var errGitHubChanged = errors.New("GitHub side effects detected")

func runGitHub(args []string, out io.Writer, client ghguard.Client) error {
	f := flag.NewFlagSet(args[0], flag.ContinueOnError)
	f.SetOutput(io.Discard)
	if args[0] == "ghsnap" {
		path := f.String("out", "", "new snapshot JSON file")
		repos := f.String("repos", "", "explicit comma-separated owner/repo scope")
		if err := f.Parse(args[1:]); err != nil {
			return err
		}
		if *path == "" || f.NArg() != 0 {
			return fmt.Errorf("usage: lb ghsnap --out FILE.json [--repos owner/repo,...]")
		}
		var names []string
		if *repos != "" {
			for _, name := range strings.Split(*repos, ",") {
				names = append(names, strings.TrimSpace(name))
			}
		}
		s, err := client.Snapshot(names)
		if err != nil {
			return err
		}
		if err := ghguard.Save(*path, s); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Snapshotted %d repositories (%s) to %s\n", len(s.Repos), s.Scope, *path)
		return err
	}
	sinceText := f.String("since", "", "include repository events after RFC3339 timestamp")
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if f.NArg() != 2 {
		return fmt.Errorf("usage: lb ghdiff [--since RFC3339] BEFORE.json AFTER.json")
	}
	var since time.Time
	if *sinceText != "" {
		var err error
		since, err = time.Parse(time.RFC3339, *sinceText)
		if err != nil {
			return fmt.Errorf("--since: %w", err)
		}
	}
	before, err := ghguard.Load(f.Arg(0))
	if err != nil {
		return fmt.Errorf("before snapshot: %w", err)
	}
	after, err := ghguard.Load(f.Arg(1))
	if err != nil {
		return fmt.Errorf("after snapshot: %w", err)
	}
	changes := ghguard.Diff(before, after)
	var eventErr error
	if !since.IsZero() {
		var events []string
		events, eventErr = client.Events(ghguard.RepoNames(before, after), since)
		changes = append(changes, events...)
	}
	for _, change := range changes {
		if _, err := fmt.Fprintln(out, change); err != nil {
			return err
		}
	}
	if eventErr != nil {
		return fmt.Errorf("GitHub event coverage incomplete: %w", eventErr)
	}
	if len(changes) != 0 {
		return errGitHubChanged
	}
	_, err = fmt.Fprintln(out, "no GitHub side effects")
	return err
}
