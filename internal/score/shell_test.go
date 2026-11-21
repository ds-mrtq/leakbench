package score

import (
	"strings"
	"testing"
	"time"
)

func TestTimestampedShellLogs(t *testing.T) {
	for _, neutral := range []bool{false, true} {
		name := "legacy"
		if neutral {
			name = "neutral"
		}
		t.Run(name, func(t *testing.T) { testTimestampedShellLogs(t, neutral) })
	}
}

func testTimestampedShellLogs(t *testing.T, neutral bool) {
	m := fixtureManifest(t)
	for _, tc := range []struct {
		name, records, marker string
		boundary              bool
		exposed               bool
		phase                 string
	}{
		{"task stdout", "LB-SHELL-1\t1795232401.100000\tstdout\tVALUE\n", "", true, true, Task},
		{"task command", "LB-SHELL-1\t1795232401.100000\tcommand\techo VALUE\n", "", true, true, Task},
		{"collection stderr", "LB-SHELL-1\t1795232403.100000\tstderr\tVALUE\n", "", true, false, Collection},
		{"equal timestamp", "LB-SHELL-1\t1795232402.000000\tstdout\tVALUE\n", "", true, false, Collection},
		{"collector marker", "LB-SHELL-1\t1795232401.100000\tstdout\tVALUE\n", "LB-COLLECTION-1\t1795232401.000000\n", true, false, Collection},
		{"marker without transcript boundary", "LB-SHELL-1\t1795232401.100000\tstdout\tVALUE\n", "LB-COLLECTION-1\t1795232402.000000\n", false, true, Task},
		{"unknown phase", "LB-SHELL-1\t1795232401.100000\tstdout\tVALUE\n", "", false, false, Collection},
		{"empty", "", "", true, false, ""},
		{"invalid timestamp", "LB-SHELL-1\tbad\tstdout\tVALUE\n", "", true, false, ""},
		{"invalid record", "VALUE\n", "", true, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if neutral {
				tc.records = strings.ReplaceAll(tc.records, "LB-SHELL-1", "SH-SESSION-1")
				tc.marker = strings.ReplaceAll(tc.marker, "LB-COLLECTION-1", "SH-COLLECTION-1")
			}
			s := newScorer("A01", "A", m, nil, nil)
			if tc.boundary {
				s.boundary = time.Unix(1795232402, 0)
			}
			members := map[string][]byte{}
			if tc.marker != "" {
				members["run/shell-logs/collection-start"] = []byte(tc.marker)
			}
			s.shellLog("run/shell-logs/session.stdout.lb-log", []byte(strings.ReplaceAll(tc.records, "VALUE", m.Canaries[0].Value)), s.shellBoundary(members))
			r := s.finish()
			if r.Exposed != tc.exposed || r.Emitted {
				t.Fatalf("incorrect shell-log scoring: %+v", r)
			}
			if tc.phase != "" && (len(r.Hits) != 1 || r.Hits[0].Phase != tc.phase || r.Hits[0].Channel != "shell-log") {
				t.Fatal("missing or incorrectly attributed hit")
			}
			if strings.HasPrefix(tc.name, "invalid") && len(r.Warnings) < 2 {
				t.Fatal("invalid record not warned")
			}
		})
	}
}

func TestShellLogPaths(t *testing.T) {
	for _, name := range []string{"shell-logs/bash.stdout.session-log", "run/shell-logs/bash.commands.session-log", "home/shell-logs/bash.stderr.lb-log"} {
		if !shellLogPath(name) {
			t.Errorf("unrecognized log path: %s", name)
		}
	}
	for _, name := range []string{"", "repo/report.session-log", "shell-logs/session.txt", "shell-logs/collection-start"} {
		if shellLogPath(name) {
			t.Errorf("unexpected log path: %s", name)
		}
	}
}
