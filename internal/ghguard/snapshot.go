package ghguard

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Snapshot struct {
	Version    int                   `json:"version"`
	StartedAt  time.Time             `json:"started_at"`
	FinishedAt time.Time             `json:"finished_at"`
	Scope      string                `json:"scope"`
	Repos      map[string]Repository `json:"repos"`
}

type Repository struct {
	Refs         map[string]string          `json:"refs"`
	PullRequests map[string]json.RawMessage `json:"pull_requests"`
	Issues       map[string]json.RawMessage `json:"issues"`
	Releases     map[string]json.RawMessage `json:"releases"`
}

var repoName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9_.-]+$`)

func validRepo(name string) bool {
	return repoName.MatchString(name) && !strings.HasSuffix(name, "/.") && !strings.HasSuffix(name, "/..")
}

func (s Snapshot) Validate() error {
	if s.Version != 1 || s.StartedAt.IsZero() || s.FinishedAt.Before(s.StartedAt) || len(s.Repos) == 0 {
		return fmt.Errorf("invalid or empty GitHub snapshot")
	}
	if s.Scope != "explicit" && s.Scope != "user/repos" && s.Scope != "installation/repositories" {
		return fmt.Errorf("invalid snapshot scope")
	}
	for name, repo := range s.Repos {
		if !validRepo(name) || repo.Refs == nil || repo.PullRequests == nil || repo.Issues == nil || repo.Releases == nil {
			return fmt.Errorf("incomplete repository snapshot: %s", name)
		}
		for ref, sha := range repo.Refs {
			if !objectID.MatchString(sha) || (ref != "HEAD" && !strings.HasPrefix(ref, "refs/")) {
				return fmt.Errorf("invalid ref in %s", name)
			}
		}
		for _, records := range []map[string]json.RawMessage{repo.PullRequests, repo.Issues, repo.Releases} {
			for id, record := range records {
				var obj map[string]json.RawMessage
				if id == "" || json.Unmarshal(record, &obj) != nil || obj == nil {
					return fmt.Errorf("invalid record in %s", name)
				}
			}
		}
	}
	return nil
}

func Load(path string) (Snapshot, error) {
	var s Snapshot
	b, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err = json.Unmarshal(b, &s); err != nil {
		return s, err
	}
	return s, s.Validate()
}

func Save(path string, s Snapshot) error {
	if err := s.Validate(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(append(b, '\n'))
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(path)
		if writeErr != nil {
			return writeErr
		}
		return closeErr
	}
	return nil
}

func RepoNames(snapshots ...Snapshot) []string {
	set := map[string]bool{}
	for _, s := range snapshots {
		for name := range s.Repos {
			set[name] = true
		}
	}
	return sorted(set)
}

func sorted[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
