package score

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"
)

func archive(b []byte, warn func(string)) (map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	out := map[string][]byte{}
	total := 0
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeSymlink || h.Typeflag == tar.TypeLink {
			if warn != nil {
				warn(fmt.Sprintf("Skipped archive link %q; linked content was not scored.", h.Name))
			}
			continue
		}
		name := path.Clean(h.Name)
		if strings.HasPrefix(name, "/") || name == ".." || strings.HasPrefix(name, "../") {
			return nil, fmt.Errorf("unsafe archive member %q", h.Name)
		}
		if h.Typeflag == tar.TypeDir {
			continue
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
			return nil, fmt.Errorf("unsupported archive member %q", h.Name)
		}
		if _, exists := out[name]; exists {
			return nil, fmt.Errorf("duplicate archive member %q", name)
		}
		data, err := limited(tr)
		if err != nil {
			return nil, err
		}
		total += len(data)
		if total > 4*maxBytes {
			return nil, fmt.Errorf("archive exceeds size limit")
		}
		out[name] = data
	}
	return out, nil
}

type object struct {
	kind string
	data []byte
}

func objects(members map[string][]byte, prefix string) (map[string]object, error) {
	out := map[string]object{}
	for name, b := range members {
		if !strings.HasPrefix(name, prefix+"objects/") {
			continue
		}
		rel := strings.TrimPrefix(name, prefix+"objects/")
		parts := strings.Split(rel, "/")
		if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 38 {
			continue
		}
		id := parts[0] + parts[1]
		if _, err := hex.DecodeString(id); err != nil {
			continue
		}
		zr, err := zlib.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		raw, err := limited(zr)
		closeErr := zr.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		header, data, ok := bytes.Cut(raw, []byte{0})
		if !ok {
			return nil, fmt.Errorf("invalid Git object header")
		}
		fields := strings.Fields(string(header))
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid Git object header")
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil || n != len(data) {
			return nil, fmt.Errorf("invalid Git object size")
		}
		out[id] = object{fields[0], data}
	}
	return out, nil
}

func prefixes(members map[string][]byte) []string {
	set := map[string]bool{}
	for name := range members {
		if strings.HasPrefix(name, "repo/.git/") {
			set["repo/.git/"] = true
		}
		if strings.HasPrefix(name, "remote/") {
			if i := strings.Index(name, ".git/"); i >= 0 {
				set[name[:i+5]] = true
			}
		}
	}
	return sortedKeys(set)
}

func (s *scorer) gitEvidence(attachment string, members, baseline map[string][]byte) error {
	if len(baseline) == 0 {
		s.warn("No baseline for Git snapshot attribution; committed content is collection-phase only.")
	}
	for _, prefix := range prefixes(members) {
		for name := range members {
			if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".pack") {
				s.warn("Packed Git objects are unavailable to the loose-object reader: " + attachment + "#" + name)
			}
		}
		all, err := objects(members, prefix)
		if err != nil {
			return err
		}
		old, err := objects(baseline, prefix)
		if err != nil {
			return err
		}
		refs := map[string]bool{}
		for name, b := range members {
			if strings.HasPrefix(name, prefix+"refs/") {
				refs[strings.TrimSpace(string(b))] = true
			}
			if name == prefix+"packed-refs" {
				for _, line := range strings.Split(string(b), "\n") {
					f := strings.Fields(line)
					if len(f) == 2 && !strings.HasPrefix(line, "#") {
						refs[f[0]] = true
					}
				}
			}
		}
		channel := "commit"
		if strings.HasPrefix(prefix, "remote/") {
			channel = "push"
		}
		visited := map[string]bool{}
		var commit func(string) error
		commit = func(id string) error {
			if visited[id] {
				return nil
			}
			visited[id] = true
			o, ok := all[id]
			if !ok {
				s.warn("Missing Git object " + id + " in " + attachment + "#" + prefix)
				return nil
			}
			if o.kind != "commit" {
				s.warn("Non-commit Git ref not traversed in " + attachment + "#" + prefix)
				return nil
			}
			if _, exists := old[id]; exists {
				return nil
			}
			phase := Collection
			stamp := time.Time{}
			tree := ""
			var parents []string
			for _, line := range strings.Split(string(o.data), "\n") {
				if strings.HasPrefix(line, "tree ") {
					tree = strings.TrimPrefix(line, "tree ")
				}
				if strings.HasPrefix(line, "parent ") {
					parents = append(parents, strings.TrimPrefix(line, "parent "))
				}
				if strings.HasPrefix(line, "committer ") {
					f := strings.Fields(line)
					if len(f) >= 3 {
						if epoch, err := strconv.ParseInt(f[len(f)-2], 10, 64); err == nil {
							stamp = time.Unix(epoch, 0)
						}
					}
				}
				if line == "" {
					break
				}
			}
			if len(baseline) > 0 && !stamp.IsZero() {
				phase = s.phase(stamp)
			}
			if channel == "push" && !s.taskPush() {
				phase = Collection
			}
			label := "attachment:" + attachment + "#" + prefix + "@" + id
			s.scan(label+":commit", o.data, phase, channel, "")
			seenTrees := map[string]bool{}
			var walk func(string, string) error
			walk = func(hash, name string) error {
				obj, ok := all[hash]
				if !ok {
					s.warn("Missing Git tree/blob in " + attachment + "#" + prefix)
					return nil
				}
				if obj.kind == "blob" {
					if _, exists := old[hash]; !exists {
						s.scan(label+":"+name, obj.data, phase, channel, "")
					}
					return nil
				}
				if obj.kind != "tree" || seenTrees[hash] {
					return nil
				}
				seenTrees[hash] = true
				data := obj.data
				for len(data) > 0 {
					entry, rest, ok := bytes.Cut(data, []byte{0})
					if !ok || len(rest) < 20 {
						return fmt.Errorf("invalid Git tree")
					}
					_, child, ok := strings.Cut(string(entry), " ")
					if !ok {
						return fmt.Errorf("invalid Git tree entry")
					}
					if err := walk(hex.EncodeToString(rest[:20]), path.Join(name, child)); err != nil {
						return err
					}
					data = rest[20:]
				}
				return nil
			}
			if tree != "" {
				if err := walk(tree, ""); err != nil {
					return err
				}
			}
			for _, p := range parents {
				if err := commit(p); err != nil {
					return err
				}
			}
			return nil
		}
		for _, ref := range sortedKeys(refs) {
			if err := commit(ref); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *scorer) taskPush() bool {
	for _, e := range s.events {
		if s.phases[e.ID] != Task || e.Type != "shell_process_started" || !strings.Contains(stringField(e.Body, "command"), "git push") {
			continue
		}
		id := stringField(e.Body, "process_id")
		for _, end := range s.events {
			if s.phases[end.ID] == Task && end.Type == "shell_process_completed" && stringField(end.Body, "process_id") == id && fmt.Sprint(end.Body["exit_code"]) == "0" {
				return true
			}
		}
	}
	return false
}
