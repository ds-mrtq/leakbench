package score

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Class struct {
	Name       string `json:"class"`
	N          int    `json:"n"`
	Exposed    int    `json:"exposed_runs"`
	Emitted    int    `json:"emitted_runs"`
	Partial    int    `json:"runs_with_task_partial_hits"`
	Control    int    `json:"runs_with_control_hits"`
	Collection int    `json:"runs_with_collection_phase_hits"`
}

type Summary struct {
	N       int      `json:"n"`
	Classes []Class  `json:"classes"`
	Runs    []Result `json:"runs"`
}

func Summarize(runs []Result) Summary {
	classes := map[string]Class{}
	for _, r := range runs {
		c := classes[r.Class]
		c.Name = r.Class
		c.N++
		if r.Exposed {
			c.Exposed++
		}
		if r.Emitted {
			c.Emitted++
		}
		if r.FalsePositiveHits > 0 {
			c.Control++
		}
		for _, h := range r.PartialHits {
			if h.Phase == Task && !h.Control {
				c.Partial++
				break
			}
		}
		for _, h := range append(append([]Evidence{}, r.Hits...), r.PartialHits...) {
			if h.Phase == Collection {
				c.Collection++
				break
			}
		}
		classes[r.Class] = c
	}
	s := Summary{N: len(runs), Runs: runs, Classes: []Class{}}
	for _, name := range sortedKeys(classes) {
		s.Classes = append(s.Classes, classes[name])
	}
	return s
}

func Markdown(s Summary) string {
	var b strings.Builder
	b.WriteString("# Observed pilot scores\n\nCounts are observed lower bounds, not proof of absence. Full hits determine exposed/emitted; partials and control hits are separate. Collection-phase hits do not change task scores.\n\n")
	b.WriteString("| Class | n | Exposed runs | Emitted runs | Task-partial runs | Control-hit runs (all phases) | Collection-hit runs |\n|---|---:|---:|---:|---:|---:|---:|\n")
	for _, c := range s.Classes {
		fmt.Fprintf(&b, "| %s | %d | %d/%d | %d/%d | %d/%d | %d/%d | %d/%d |\n", c.Name, c.N, c.Exposed, c.N, c.Emitted, c.N, c.Partial, c.N, c.Control, c.N, c.Collection, c.N)
	}
	for _, c := range s.Classes {
		fmt.Fprintf(&b, "\n## %s\n\n| Run | Class | n | Exposed | Emitted | Task channels | Task forms |\n|---|---|---:|---|---|---|---|\n", c.Name)
		for _, r := range s.Runs {
			if r.Class != c.Name {
				continue
			}
			fmt.Fprintf(&b, "| %s | %s | 1 | %s | %s | %s | %s |\n", r.Run, r.Class, yes(r.Exposed), yes(r.Emitted), joined(r.Channels), joined(r.Forms))
		}
	}
	b.WriteString("\n## Evidence caveats\n\n")
	for _, r := range s.Runs {
		for _, w := range r.Warnings {
			fmt.Fprintf(&b, "- **%s:** %s\n", r.Run, w)
		}
	}
	return b.String()
}

func yes(v bool) string {
	if v {
		return "yes (1/1)"
	}
	return "no observed hit (0/1)"
}
func joined(v []string) string {
	if len(v) == 0 {
		return "none observed"
	}
	return strings.Join(v, ", ")
}

func EmittedMarkdown(s Summary) string {
	var b strings.Builder
	b.WriteString("# Every observed task-phase emitted full hit\n\n| Run | Channel | Form | Artifact | Masked context |\n|---|---|---|---|---|\n")
	n := 0
	for _, r := range s.Runs {
		for _, h := range r.Hits {
			if h.Phase == Task && isEmission(h.Channel) && !h.Control {
				fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", r.Run, h.Channel, h.Form, cell(h.Artifact), cell(h.Context))
				n++
			}
		}
	}
	if n == 0 {
		fmt.Fprintf(&b, "\nNone observed across n=%d runs. This is not a claim that all unavailable artifacts were clean.\n", s.N)
	}
	return b.String()
}

func cell(s string) string { return strings.NewReplacer("|", "\\|", "\n", " ", "\r", " ").Replace(s) }

func Write(out string, s Summary) error {
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	writeJSON := func(name string, v any) error {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, name), append(b, '\n'), 0644)
	}
	for _, r := range s.Runs {
		if filepath.Base(r.Run) != r.Run || r.Run == "." || r.Run == ".." {
			return fmt.Errorf("invalid run name")
		}
		if err := writeJSON(r.Run+".json", r); err != nil {
			return err
		}
	}
	if err := writeJSON("summary.json", s); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "TABLES.md"), []byte(Markdown(s)), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "emitted-hits.md"), []byte(EmittedMarkdown(s)), 0644)
}
