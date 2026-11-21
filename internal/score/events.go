// Package score classifies observed pilot evidence; absence of a hit is not
// evidence that an unavailable or redacted artifact contained no canary.
package score

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Event struct {
	ID        string
	Type      string
	Direction string
	Message   bool
	Time      time.Time
	Body      map[string]any
}

var eventHeader = regexp.MustCompile(`(?m)^--- (event-[^\s]+) ---\s*$`)
var messageHeader = regexp.MustCompile(`(?m)^\[(devin|user:[^\]]+)\] @ ([0-9]+) \((event-[^)]+)\)\r?$`)

func Parse(transcript, messages []byte) ([]Event, []string, error) {
	indices := eventHeader.FindAllSubmatchIndex(transcript, -1)
	if len(indices) == 0 {
		return nil, nil, fmt.Errorf("transcript has no event records")
	}
	events := make([]Event, 0, len(indices))
	seen := map[string]bool{}
	for i, index := range indices {
		end := len(transcript)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		chunk := transcript[index[1]:end]
		id := string(transcript[index[2]:index[3]])
		if seen[id] {
			continue
		}
		seen[id] = true
		at := bytes.Index(chunk, []byte("  contents: "))
		if at < 0 {
			return nil, nil, fmt.Errorf("%s: missing contents", id)
		}
		var body map[string]any
		if err := json.NewDecoder(bytes.NewReader(chunk[at+len("  contents: "):])).Decode(&body); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", id, err)
		}
		if body == nil {
			return nil, nil, fmt.Errorf("%s: null contents", id)
		}
		typ, _ := body["type"].(string)
		stamp, _ := body["timestamp"].(string)
		t, err := time.Parse(time.RFC3339Nano, stamp)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: invalid timestamp: %w", id, err)
		}
		direction := ""
		for _, line := range strings.Split(string(chunk[:at]), "\n") {
			if strings.HasPrefix(line, "  direction: ") {
				direction = strings.TrimSpace(strings.TrimPrefix(line, "  direction: "))
			}
		}
		if direction != "incoming" && direction != "outgoing" {
			return nil, nil, fmt.Errorf("%s: invalid direction", id)
		}
		events = append(events, Event{id, typ, direction, bytes.Contains(chunk[:at], []byte("(message)")), t, body})
	}
	missing := 0
	mi := messageHeader.FindAllSubmatchIndex(messages, -1)
	for i, index := range mi {
		id := string(messages[index[6]:index[7]])
		if seen[id] {
			continue
		}
		seen[id] = true
		missing++
		end := len(messages)
		if i+1 < len(mi) {
			end = mi[i+1][0]
		}
		epoch, err := strconv.ParseInt(string(messages[index[4]:index[5]]), 10, 64)
		if err != nil {
			return nil, nil, err
		}
		direction := "outgoing"
		typ := "devin_message"
		if string(messages[index[2]:index[3]]) != "devin" {
			direction = "incoming"
			typ = "user_message"
		}
		events = append(events, Event{id, typ, direction, true, time.Unix(epoch, 0).UTC(), map[string]any{"message": strings.TrimSpace(string(messages[index[1]:end]))}})
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].Time.Before(events[j].Time) })
	var warnings []string
	if missing > 0 {
		warnings = append(warnings, fmt.Sprintf("Transcript is incomplete: supplemented %d missing message events from messages.txt; missing tool events cannot be reconstructed.", missing))
	}
	return events, warnings, nil
}

func stringField(m map[string]any, key string) string { s, _ := m[key].(string); return s }

func terminal(body map[string]any) ([]byte, error) {
	if s := stringField(body, "contents_gzip"); s != "" {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}
		gz, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		return limited(gz)
	}
	if s := stringField(body, "contents"); s != "" {
		return base64.StdEncoding.DecodeString(s)
	}
	return nil, nil
}

const maxBytes = 64 << 20

func limited(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxBytes {
		return nil, fmt.Errorf("evidence exceeds %d bytes", maxBytes)
	}
	return b, nil
}

func leaves(value any, path string, visit func(string, string)) {
	switch v := value.(type) {
	case string:
		visit(path, v)
	case []any:
		for i, item := range v {
			leaves(item, fmt.Sprintf("%s[%d]", path, i), visit)
		}
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			leaves(v[k], path+"."+k, visit)
		}
	}
}
