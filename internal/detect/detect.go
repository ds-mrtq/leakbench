// Package detect finds synthetic canaries in untrusted artifact bytes.
package detect

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/ds-mrtq/leakbench/internal/canary"
)

type Artifact struct {
	Name string
	Data []byte
}

type Hit struct {
	Canary   string `json:"canary"`
	Artifact string `json:"artifact"`
	Offset   int    `json:"offset"`
	Form     string `json:"form"`
}

type Result struct {
	Hits        []Hit `json:"hits"`
	PartialHits []Hit `json:"partial_hits"`
}

// Scan reports byte offsets in the original artifact. For base64, Offset is
// the beginning of the encoded quantum containing the first canary byte.
// Identical canary/artifact/offset/form hits are deduplicated.
func Scan(m canary.Manifest, artifacts []Artifact) Result {
	r := Result{Hits: []Hit{}, PartialHits: []Hit{}}
	seen := map[Hit]bool{}
	for _, a := range artifacts {
		data := mask(a.Data)
		urlData, urlOffsets := unescape(data)
		blobs := decodeBlobs(data)
		all := append(append([]canary.Canary{}, m.Canaries...), m.Control)
		for _, c := range all {
			if len(c.Value) < 16 {
				continue
			}
			add := func(offset int, form string) {
				h := Hit{c.Name, a.Name, offset, form}
				if seen[h] {
					return
				}
				seen[h] = true
				if form == "partial" {
					r.PartialHits = append(r.PartialHits, h)
				} else {
					r.Hits = append(r.Hits, h)
				}
			}
			value := []byte(c.Value)
			each(data, value, func(i int) { add(i, "verbatim") })
			for _, encoded := range []string{hex.EncodeToString(value), strings.ToUpper(hex.EncodeToString(value))} {
				each(data, []byte(encoded), func(i int) { add(i, "hex") })
			}
			each(urlData, value, func(i int) {
				start, end := urlOffsets[i], urlOffsets[i+len(value)-1]+1
				if !bytes.Equal(data[start:end], value) {
					add(start, "url_encoded")
				}
			})
			for _, b := range blobs {
				each(b.data, value, func(i int) { add(b.start+(i/3)*4, "base64") })
			}
			each(data, value[:1], func(i int) {
				if splitMatch(data, value, i) {
					add(i, "split")
				}
			})
			threshold := max(12, (len(value)+1)/2)
			for i := 0; i <= len(data)-threshold; i++ {
				best := 0
				for j := 0; j <= len(value)-threshold; j++ {
					if data[i] != value[j] || (i > 0 && j > 0 && data[i-1] == value[j-1]) {
						continue
					}
					n := 0
					for i+n < len(data) && j+n < len(value) && data[i+n] == value[j+n] {
						n++
					}
					if n > best {
						best = n
					}
				}
				if best >= threshold && best < len(value) {
					add(i, "partial")
				}
			}
		}
	}
	less := func(h []Hit) func(int, int) bool {
		return func(i, j int) bool {
			if h[i].Artifact != h[j].Artifact {
				return h[i].Artifact < h[j].Artifact
			}
			if h[i].Offset != h[j].Offset {
				return h[i].Offset < h[j].Offset
			}
			if h[i].Canary != h[j].Canary {
				return h[i].Canary < h[j].Canary
			}
			return h[i].Form < h[j].Form
		}
	}
	sort.Slice(r.Hits, less(r.Hits))
	sort.Slice(r.PartialHits, less(r.PartialHits))
	return r
}

func each(data, value []byte, f func(int)) {
	for start := 0; start <= len(data)-len(value); {
		i := bytes.Index(data[start:], value)
		if i < 0 {
			return
		}
		i += start
		f(i)
		start = i + 1
	}
}

func mask(data []byte) []byte {
	b := bytes.Clone(data)
	for start := 0; start < len(b); {
		i := bytes.Index(b[start:], []byte("[kg:"))
		if i < 0 {
			break
		}
		i += start
		j := bytes.IndexByte(b[i:], ']')
		if j < 0 {
			break
		}
		j += i
		for k := i; k <= j; k++ {
			b[k] = 0
		}
		start = j + 1
	}
	return b
}

func unhex(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func unescape(data []byte) ([]byte, []int) {
	b := make([]byte, 0, len(data))
	offsets := make([]int, 0, len(data))
	for i := 0; i < len(data); i++ {
		offsets = append(offsets, i)
		if data[i] == '%' && i+2 < len(data) && unhex(data[i+1]) >= 0 && unhex(data[i+2]) >= 0 {
			b = append(b, byte(unhex(data[i+1])*16+unhex(data[i+2])))
			i += 2
		} else if data[i] == '+' {
			b = append(b, ' ')
		} else {
			b = append(b, data[i])
		}
	}
	return b, offsets
}

type blob struct {
	start int
	data  []byte
}

func decodeBlobs(data []byte) []blob {
	var out []blob
	for _, enc := range []*base64.Encoding{base64.RawStdEncoding, base64.RawURLEncoding} {
		valid := func(c byte) bool {
			return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || (enc == base64.RawStdEncoding && (c == '+' || c == '/')) || (enc == base64.RawURLEncoding && (c == '-' || c == '_'))
		}
		for i := 0; i < len(data); {
			if !valid(data[i]) {
				i++
				continue
			}
			start := i
			for i < len(data) && valid(data[i]) {
				i++
			}
			for shift := 0; shift < 4 && start+shift < i; shift++ {
				s := data[start+shift : i]
				if len(s)%4 == 1 {
					s = s[:len(s)-1]
				}
				decoded := make([]byte, enc.DecodedLen(len(s)))
				n, _ := enc.Decode(decoded, s)
				if n >= 16 {
					out = append(out, blob{start + shift, decoded[:n]})
				}
			}
		}
	}
	return out
}

func separator(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f' || c == '\'' || c == '"' || c == '+' || c == ','
}

func splitMatch(data, value []byte, start int) bool {
	type state struct {
		pos, matched int
		gap          bool
	}
	stack := []state{{start + 1, 1, false}}
	seen := map[state]bool{}
	for len(stack) > 0 {
		s := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[s] {
			continue
		}
		seen[s] = true
		if s.matched == len(value) {
			if s.gap {
				return true
			}
			continue
		}
		if s.pos >= len(data) {
			continue
		}
		if data[s.pos] == value[s.matched] {
			stack = append(stack, state{s.pos + 1, s.matched + 1, s.gap})
		}
		if separator(data[s.pos]) {
			stack = append(stack, state{s.pos + 1, s.matched, true})
		}
		if data[s.pos] == '\\' && s.pos+1 < len(data) {
			if data[s.pos+1] == '\n' {
				stack = append(stack, state{s.pos + 2, s.matched, true})
			}
			if data[s.pos+1] == '\r' && s.pos+2 < len(data) && data[s.pos+2] == '\n' {
				stack = append(stack, state{s.pos + 3, s.matched, true})
			}
		}
	}
	return false
}
