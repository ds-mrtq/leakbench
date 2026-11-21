package score

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func logTime(text string) (time.Time, error) {
	seconds, fraction, ok := strings.Cut(text, ".")
	if !ok || len(fraction) == 0 || len(fraction) > 9 {
		return time.Time{}, fmt.Errorf("invalid shell-log timestamp")
	}
	for _, c := range seconds + fraction {
		if c < '0' || c > '9' {
			return time.Time{}, fmt.Errorf("invalid shell-log timestamp")
		}
	}
	sec, err := strconv.ParseInt(seconds, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	nanos, err := strconv.ParseInt(fraction+strings.Repeat("0", 9-len(fraction)), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, nanos), nil
}

func shellPath(name string) bool {
	return strings.HasPrefix(name, "shell-logs/") || strings.Contains(name, "/shell-logs/")
}

func (s *scorer) shellBoundary(members map[string][]byte) time.Time {
	boundary := s.boundary
	for name, b := range members {
		if !shellPath(name) || !strings.HasSuffix(name, "/collection-start") {
			continue
		}
		version, stamp, ok := strings.Cut(strings.TrimSpace(string(b)), "\t")
		t, err := logTime(stamp)
		if !ok || version != "LB-COLLECTION-1" || err != nil {
			s.warn("Invalid shell collection marker: " + name)
			continue
		}
		if boundary.IsZero() || t.Before(boundary) {
			boundary = t
		}
	}
	return boundary
}

func (s *scorer) shellLog(name string, b []byte, boundary time.Time) {
	var task, collection bytes.Buffer
	if boundary.IsZero() {
		s.warn("No collection boundary for shell logs; records are collection-phase only.")
	}
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		fields := bytes.SplitN(line, []byte{'\t'}, 4)
		if len(fields) != 4 || string(fields[0]) != "LB-SHELL-1" {
			s.warn("Invalid shell-log record skipped: " + name)
			continue
		}
		stamp, err := logTime(string(fields[1]))
		kind := string(fields[2])
		if err != nil || (kind != "command" && kind != "stdout" && kind != "stderr") {
			s.warn("Invalid shell-log record skipped: " + name)
			continue
		}
		out := &collection
		if !boundary.IsZero() && stamp.Before(boundary) {
			out = &task
		}
		out.Write(fields[3])
		out.WriteByte('\n')
	}
	s.scan(name+":task", task.Bytes(), Task, "shell-log", "")
	s.scan(name+":collection", collection.Bytes(), Collection, "shell-log", "")
}
