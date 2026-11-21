package ghguard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

func canonical(b json.RawMessage) string {
	var value any
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if d.Decode(&value) != nil {
		return string(b)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return string(b)
	}
	return string(encoded)
}

func Diff(before, after Snapshot) []string {
	var changes []string
	for _, name := range RepoNames(before, after) {
		old, had := before.Repos[name]
		current, has := after.Repos[name]
		if !had {
			changes = append(changes, name+": repository added to snapshot scope")
			continue
		}
		if !has {
			changes = append(changes, name+": repository removed from snapshot scope")
			continue
		}
		refs := map[string]bool{}
		for ref := range old.Refs {
			refs[ref] = true
		}
		for ref := range current.Refs {
			refs[ref] = true
		}
		for _, ref := range sorted(refs) {
			if old.Refs[ref] != current.Refs[ref] {
				changes = append(changes, fmt.Sprintf("%s: ref %s: %s -> %s", name, ref, old.Refs[ref], current.Refs[ref]))
			}
		}
		for _, group := range []struct {
			name         string
			old, current map[string]json.RawMessage
		}{
			{"pull request", old.PullRequests, current.PullRequests},
			{"issue", old.Issues, current.Issues},
			{"release", old.Releases, current.Releases},
		} {
			keys := map[string]bool{}
			for key := range group.old {
				keys[key] = true
			}
			for key := range group.current {
				keys[key] = true
			}
			for _, key := range sorted(keys) {
				previous, had := group.old[key]
				next, has := group.current[key]
				action := "changed"
				if !had {
					action = "added"
				} else if !has {
					action = "removed"
				} else if canonical(previous) == canonical(next) {
					continue
				}
				changes = append(changes, fmt.Sprintf("%s: %s #%s %s", name, group.name, key, action))
			}
		}
	}
	return changes
}

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Payload   json.RawMessage `json:"payload"`
	Actor     struct {
		Login string `json:"login"`
	} `json:"actor"`
}

func EventDiff(repo string, events []Event, since time.Time) []string {
	events = append([]Event(nil), events...)
	allowed := map[string]bool{"PushEvent": true, "CreateEvent": true, "DeleteEvent": true, "PullRequestEvent": true, "IssuesEvent": true, "IssueCommentEvent": true, "ReleaseEvent": true}
	var changes []string
	seen := map[string]bool{}
	sort.Slice(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].ID < events[j].ID
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	for _, event := range events {
		if !allowed[event.Type] || !event.CreatedAt.After(since) || seen[event.ID] {
			continue
		}
		seen[event.ID] = true
		var payload struct {
			Action string `json:"action"`
			Ref    string `json:"ref"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		changes = append(changes, fmt.Sprintf("%s: event %s %s at %s actor=%q action=%q ref=%q", repo, event.ID, event.Type, event.CreatedAt.UTC().Format(time.RFC3339Nano), event.Actor.Login, payload.Action, payload.Ref))
	}
	return changes
}

func (c Client) Events(names []string, since time.Time) ([]string, error) {
	var changes []string
	for _, name := range names {
		if !validRepo(name) {
			return changes, fmt.Errorf("invalid repository %q", name)
		}
		rows, err := c.list("repos/"+name+"/events?per_page=100", "")
		if err != nil {
			return changes, err
		}
		var events []Event
		var oldest time.Time
		for _, row := range rows {
			var event Event
			if err := json.Unmarshal(row, &event); err != nil {
				return changes, err
			}
			if event.ID == "" || event.CreatedAt.IsZero() {
				return changes, fmt.Errorf("invalid event for %s", name)
			}
			if oldest.IsZero() || event.CreatedAt.Before(oldest) {
				oldest = event.CreatedAt
			}
			events = append(events, event)
		}
		changes = append(changes, EventDiff(name, events, since)...)
		if len(rows) >= 300 && oldest.After(since) {
			return changes, fmt.Errorf("%s: event feed cap reached after --since; history coverage incomplete", name)
		}
	}
	return changes, nil
}
