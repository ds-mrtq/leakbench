package ghguard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Runner func(string, ...string) ([]byte, error)

type Client struct {
	Run Runner
}

func Execute(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GH_PROMPT_DISABLED=1")
	cmd.WaitDelay = 2 * time.Second
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return b, nil
}

func (c Client) command(name string, args ...string) ([]byte, error) {
	if c.Run != nil {
		return c.Run(name, args...)
	}
	return Execute(name, args...)
}

func (c Client) list(endpoint, field string) ([]json.RawMessage, error) {
	b, err := c.command("gh", "api", "--hostname", "github.com", "--method", "GET", "--paginate", endpoint)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", endpoint, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	var all []json.RawMessage
	pages := 0
	for {
		var page json.RawMessage
		if err := decoder.Decode(&page); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("GET %s: invalid JSON: %w", endpoint, err)
		}
		pages++
		if field != "" {
			var object map[string]json.RawMessage
			if err := json.Unmarshal(page, &object); err != nil {
				return nil, err
			}
			page = object[field]
		}
		if len(bytes.TrimSpace(page)) == 0 || bytes.TrimSpace(page)[0] != '[' {
			return nil, fmt.Errorf("GET %s: expected array page", endpoint)
		}
		var rows []json.RawMessage
		if err := json.Unmarshal(page, &rows); err != nil {
			return nil, err
		}
		all = append(all, rows...)
	}
	if pages == 0 {
		return nil, fmt.Errorf("GET %s: empty response", endpoint)
	}
	return all, nil
}

func (c Client) repositories(explicit []string) ([]string, string, error) {
	if len(explicit) != 0 {
		set := map[string]bool{}
		for _, name := range explicit {
			if !validRepo(name) {
				return nil, "", fmt.Errorf("invalid repository %q; use owner/repo", name)
			}
			set[name] = true
		}
		return sorted(set), "explicit", nil
	}
	var failures []string
	for _, source := range []string{"user/repos", "installation/repositories"} {
		field := ""
		if source == "installation/repositories" {
			field = "repositories"
		}
		rows, err := c.list(source+"?per_page=100", field)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		set := map[string]bool{}
		for _, row := range rows {
			var repo struct {
				FullName string `json:"full_name"`
			}
			if err := json.Unmarshal(row, &repo); err != nil || !validRepo(repo.FullName) {
				return nil, "", fmt.Errorf("invalid repository returned by %s", source)
			}
			set[repo.FullName] = true
		}
		if len(set) != 0 {
			return sorted(set), source, nil
		}
		failures = append(failures, source+": no visible repositories")
	}
	return nil, "", fmt.Errorf("repository discovery failed; supply --repos owner/repo,...: %s", strings.Join(failures, "; "))
}

var objectID = regexp.MustCompile(`^(?:[a-fA-F0-9]{40}|[a-fA-F0-9]{64})$`)

func parseRefs(b []byte) (map[string]string, error) {
	refs := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || !objectID.MatchString(fields[0]) || (fields[1] != "HEAD" && !strings.HasPrefix(fields[1], "refs/")) {
			return nil, fmt.Errorf("invalid git ls-remote record")
		}
		if _, exists := refs[fields[1]]; exists {
			return nil, fmt.Errorf("duplicate remote ref %s", fields[1])
		}
		refs[fields[1]] = fields[0]
	}
	return refs, nil
}

func index(rows []json.RawMessage, field string) (map[string]json.RawMessage, error) {
	records := map[string]json.RawMessage{}
	for _, row := range rows {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(row, &obj); err != nil {
			return nil, err
		}
		id := string(obj[field])
		if id == "" || id == "0" || id == "null" || strings.Trim(id, "0123456789") != "" {
			return nil, fmt.Errorf("invalid %s in API record", field)
		}
		if _, exists := records[id]; exists {
			return nil, fmt.Errorf("duplicate API record %s", id)
		}
		records[id] = row
	}
	return records, nil
}

func (c Client) Snapshot(explicit []string) (Snapshot, error) {
	s := Snapshot{Version: 1, StartedAt: time.Now().UTC(), Repos: map[string]Repository{}}
	names, scope, err := c.repositories(explicit)
	if err != nil {
		return s, err
	}
	s.Scope = scope
	for _, name := range names {
		b, err := c.command("git", "ls-remote", "https://github.com/"+name+".git")
		if err != nil {
			return s, fmt.Errorf("refs for %s: %w", name, err)
		}
		refs, err := parseRefs(b)
		if err != nil {
			return s, fmt.Errorf("refs for %s: %w", name, err)
		}
		repo := Repository{Refs: refs}
		for _, spec := range []struct {
			path, key string
			dest      *map[string]json.RawMessage
		}{
			{"pulls?state=open&per_page=100", "number", &repo.PullRequests},
			{"issues?state=all&per_page=100", "number", &repo.Issues},
			{"releases?per_page=100", "id", &repo.Releases},
		} {
			rows, err := c.list("repos/"+name+"/"+spec.path, "")
			if err != nil {
				return s, err
			}
			*spec.dest, err = index(rows, spec.key)
			if err != nil {
				return s, fmt.Errorf("%s %s: %w", name, spec.path, err)
			}
		}
		s.Repos[name] = repo
	}
	s.FinishedAt = time.Now().UTC()
	return s, s.Validate()
}
