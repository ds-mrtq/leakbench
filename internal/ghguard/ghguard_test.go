package ghguard

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sha = "0123456789012345678901234567890123456789"

func fixture() Snapshot {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return Snapshot{Version: 1, StartedAt: now, FinishedAt: now.Add(time.Second), Scope: "explicit", Repos: map[string]Repository{
		"owner/project": {Refs: map[string]string{"refs/heads/main": sha}, PullRequests: map[string]json.RawMessage{}, Issues: map[string]json.RawMessage{}, Releases: map[string]json.RawMessage{}},
	}}
}

func TestDiff(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		change     func(*Snapshot)
	}{
		{"unchanged", "", func(s *Snapshot) { s.FinishedAt = s.FinishedAt.Add(time.Hour) }},
		{"branch", "ref refs/heads/main", func(s *Snapshot) { s.Repos["owner/project"].Refs["refs/heads/main"] = strings.Repeat("a", 40) }},
		{"tag", "ref refs/tags/v1", func(s *Snapshot) { s.Repos["owner/project"].Refs["refs/tags/v1"] = sha }},
		{"pull ref", "ref refs/pull/4/head", func(s *Snapshot) { s.Repos["owner/project"].Refs["refs/pull/4/head"] = sha }},
		{"deleted ref", "ref refs/heads/main", func(s *Snapshot) { delete(s.Repos["owner/project"].Refs, "refs/heads/main") }},
		{"PR", "pull request #4 added", func(s *Snapshot) {
			s.Repos["owner/project"].PullRequests["4"] = json.RawMessage(`{"number":4,"state":"open"}`)
		}},
		{"issue", "issue #5 added", func(s *Snapshot) {
			s.Repos["owner/project"].Issues["5"] = json.RawMessage(`{"number":5,"comments":1,"updated_at":"today"}`)
		}},
		{"release", "release #9 added", func(s *Snapshot) {
			s.Repos["owner/project"].Releases["9"] = json.RawMessage(`{"id":9,"tag_name":"v1"}`)
		}},
		{"repo removed", "repository removed", func(s *Snapshot) { delete(s.Repos, "owner/project") }},
		{"repo added", "repository added", func(s *Snapshot) { s.Repos["owner/new"] = s.Repos["owner/project"] }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, after := fixture(), fixture()
			tc.change(&after)
			changes := Diff(before, after)
			if tc.want == "" {
				if len(changes) != 0 {
					t.Fatal(changes)
				}
				return
			}
			if len(changes) != 1 || !strings.Contains(changes[0], tc.want) {
				t.Fatalf("unexpected differences: %v", changes)
			}
		})
	}
	if len(Diff(Snapshot{}, Snapshot{})) != 0 {
		t.Fatal("empty diff")
	}
}

func TestRecordChanges(t *testing.T) {
	for _, field := range []string{"comments", "updated_at", "body"} {
		before, after := fixture(), fixture()
		before.Repos["owner/project"].Issues["1"] = json.RawMessage(`{"number":1,"comments":0,"updated_at":"old","body":"original"}`)
		var obj map[string]any
		_ = json.Unmarshal(before.Repos["owner/project"].Issues["1"], &obj)
		obj[field] = "new"
		b, _ := json.Marshal(obj)
		after.Repos["owner/project"].Issues["1"] = b
		if changes := Diff(before, after); len(changes) != 1 || !strings.Contains(changes[0], "issue #1 changed") {
			t.Fatal(changes)
		}
	}
	before, after := fixture(), fixture()
	before.Repos["owner/project"].Releases["7"] = json.RawMessage(`{"id":7, "body":"text"}`)
	after.Repos["owner/project"].Releases["7"] = json.RawMessage(`{ "body":"text", "id":7 }`)
	if len(Diff(before, after)) != 0 {
		t.Fatal("JSON formatting is not a side effect")
	}
	delete(after.Repos["owner/project"].Releases, "7")
	if changes := Diff(before, after); len(changes) != 1 || !strings.Contains(changes[0], "release #7 removed") {
		t.Fatal(changes)
	}
}

func TestTransientEvents(t *testing.T) {
	since := fixture().StartedAt
	var events []Event
	for i, kind := range []string{"PushEvent", "CreateEvent", "DeleteEvent", "PullRequestEvent", "IssuesEvent", "IssueCommentEvent", "ReleaseEvent"} {
		events = append(events, Event{ID: fmt.Sprint(i), Type: kind, CreatedAt: since.Add(time.Second), Payload: json.RawMessage(`{"ref":"temporary","action":"created"}`)})
	}
	events = append(events, events[0], Event{ID: "old", Type: "PushEvent", CreatedAt: since.Add(-time.Second)}, Event{ID: "equal", Type: "PushEvent", CreatedAt: since}, Event{ID: "other", Type: "WatchEvent", CreatedAt: since.Add(time.Second)})
	changes := EventDiff("owner/project", events, since)
	if len(changes) != 7 {
		t.Fatal(changes)
	}
	if !strings.Contains(strings.Join(changes, "\n"), "DeleteEvent") || !strings.Contains(changes[0], "temporary") {
		t.Fatal(changes)
	}
	if len(Diff(fixture(), fixture())) != 0 {
		t.Fatal("transient push/delete must be detected by events, not state")
	}
	if len(EventDiff("owner/project", nil, since)) != 0 {
		t.Fatal("empty events")
	}
}

func fakeClient(responses map[string]string) Client {
	return Client{Run: func(name string, args ...string) ([]byte, error) {
		key := args[len(args)-1]
		if name == "git" {
			key = "git"
		}
		b, ok := responses[key]
		if !ok {
			return nil, errors.New("denied: " + key)
		}
		return []byte(b), nil
	}}
}

func responses() map[string]string {
	return map[string]string{
		"installation/repositories?per_page=100": `{"repositories":[{"full_name":"owner/project"}]}`,
		"git":                                    sha + "\trefs/heads/main\n" + sha + "\trefs/tags/v1\n" + sha + "\trefs/pull/2/head\n",
		"repos/owner/project/pulls?state=open&per_page=100": `[{"number":2}]`,
		"repos/owner/project/issues?state=all&per_page=100": `[{"number":1,"comments":3,"updated_at":"now"}]` + "\n" + `[{"number":2,"comments":0,"pull_request":{}}]`,
		"repos/owner/project/releases?per_page=100":         `[{"id":8,"tag_name":"v1"}]`,
		"repos/owner/project/events?per_page=100":           `[{"id":"9","type":"DeleteEvent","created_at":"2026-01-02T03:04:06Z","payload":{"ref":"temporary"}}]`,
	}
}

func TestSnapshotPaginationAndFallback(t *testing.T) {
	for _, source := range []string{"installation/repositories", "user/repos", "explicit"} {
		t.Run(source, func(t *testing.T) {
			r := responses()
			var names []string
			if source == "user/repos" {
				r["user/repos?per_page=100"] = `[{"full_name":"owner/project"}]` + "\n[]"
			}
			if source == "explicit" {
				names = []string{"owner/project", "owner/project"}
				delete(r, "installation/repositories?per_page=100")
			}
			s, err := fakeClient(r).Snapshot(names)
			if err != nil {
				t.Fatal(err)
			}
			repo := s.Repos["owner/project"]
			if s.Scope != source || len(repo.Refs) != 3 || len(repo.Issues) != 2 || len(repo.Releases) != 1 || len(repo.PullRequests) != 1 {
				t.Fatalf("incomplete snapshot: %+v", s)
			}
			if !strings.Contains(string(repo.Issues["1"]), `"comments":3`) {
				t.Fatal("missing issue metadata")
			}
		})
	}
}

func TestClientFailures(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"installation/repositories?per_page=100", `{"repositories":[]}`},
		{"installation/repositories?per_page=100", `{"repositories":[{"full_name":"../bad"}]}`},
		{"installation/repositories?per_page=100", `{}`},
		{"installation/repositories?per_page=100", ``},
		{"git", "bad output"},
		{"git", sha + "\tHEAD\n" + sha + "\tHEAD\n"},
		{"repos/owner/project/issues?state=all&per_page=100", `{`},
		{"repos/owner/project/issues?state=all&per_page=100", `[{"number":1},{"number":1}]`},
		{"repos/owner/project/issues?state=all&per_page=100", `[{}]`},
	} {
		r := responses()
		r[tc.key] = tc.value
		if _, err := fakeClient(r).Snapshot(nil); err == nil {
			t.Fatalf("accepted incomplete response for %s: %s", tc.key, tc.value)
		}
	}
	if _, err := fakeClient(map[string]string{}).Snapshot(nil); err == nil || !strings.Contains(err.Error(), "--repos") {
		t.Fatal(err)
	}
	if _, err := fakeClient(responses()).Snapshot([]string{"owner/../bad"}); err == nil {
		t.Fatal("accepted unsafe repo")
	}
	if refs, err := parseRefs(nil); err != nil || len(refs) != 0 {
		t.Fatal("empty repository refs", err)
	}
	r := responses()
	r["git"] = ""
	r["repos/owner/project/pulls?state=open&per_page=100"] = "[]"
	r["repos/owner/project/issues?state=all&per_page=100"] = "[]"
	r["repos/owner/project/releases?per_page=100"] = "[]"
	if _, err := fakeClient(r).Snapshot(nil); err != nil {
		t.Fatal("empty repository collections", err)
	}
}

func TestEventClient(t *testing.T) {
	client := fakeClient(responses())
	changes, err := client.Events([]string{"owner/project"}, fixture().StartedAt)
	if err != nil || len(changes) != 1 {
		t.Fatal(changes, err)
	}
	for _, text := range []string{`{`, `[{"id":"bad"}]`, `null`} {
		r := responses()
		r["repos/owner/project/events?per_page=100"] = text
		if _, err := fakeClient(r).Events([]string{"owner/project"}, fixture().StartedAt); err == nil {
			t.Fatal("accepted bad events")
		}
	}
	if _, err := client.Events([]string{"owner/other"}, fixture().StartedAt); err == nil {
		t.Fatal("event failure ignored")
	}
	if _, err := client.Events([]string{"../bad"}, fixture().StartedAt); err == nil {
		t.Fatal("invalid event repo")
	}
	var events []Event
	for n := 0; n < 300; n++ {
		events = append(events, Event{ID: fmt.Sprint(n), Type: "PushEvent", CreatedAt: fixture().FinishedAt})
	}
	b, _ := json.Marshal(events)
	r := responses()
	r["repos/owner/project/events?per_page=100"] = string(b)
	if _, err := fakeClient(r).Events([]string{"owner/project"}, fixture().StartedAt); err == nil {
		t.Fatal("event cap ignored")
	}
}

func TestSnapshotFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := Save(path, fixture()); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("snapshot permissions", err)
	}
	s, err := Load(path)
	if err != nil || len(Diff(fixture(), s)) != 0 {
		t.Fatal(err)
	}
	if err := Save(path, fixture()); err == nil {
		t.Fatal("overwrote snapshot")
	}
	if _, err := Load(path + "missing"); err == nil {
		t.Fatal("missing snapshot")
	}
	if err := Save(path+"invalid", Snapshot{}); err == nil {
		t.Fatal("invalid snapshot saved")
	}
	for _, text := range []string{`{`, `null`, `{"version":99}`, `{"version":1,"repos":{}}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatal("invalid snapshot accepted")
		}
	}
	for _, change := range []func(*Snapshot){
		func(s *Snapshot) { s.Scope = "unknown" },
		func(s *Snapshot) { s.Repos["../bad"] = s.Repos["owner/project"] },
		func(s *Snapshot) { s.Repos["owner/project"] = Repository{} },
		func(s *Snapshot) { s.Repos["owner/project"].Refs["HEAD"] = "bad" },
		func(s *Snapshot) { s.Repos["owner/project"].Issues["1"] = json.RawMessage(`null`) },
	} {
		s := fixture()
		change(&s)
		if s.Validate() == nil {
			t.Fatal("invalid snapshot accepted")
		}
	}
}
