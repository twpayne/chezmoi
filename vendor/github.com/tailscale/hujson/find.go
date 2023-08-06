// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hujson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var errNotFound = fmt.Errorf("value not found")

// Find locates the value specified by the JSON pointer (see RFC 6901).
// It returns nil if the value does not exist or the pointer is invalid.
// If a JSON object has multiple members matching a given name,
// the first is returned. Object names are matched exactly,
// rather than with a case-insensitive match.
func (v *Value) Find(ptr string) *Value {
	if s, err := v.find(findState{pointer: ptr}); err == nil {
		return s.value
	}
	return nil
}

type findState struct {
	pointer string // pointer[:offset] is the current value, pointer[offset:] is the remainder
	offset  int

	parent composite // nil for root pointer
	name   string    // name into parent to obtain current value
	idx    int       // idx into parent to obtain current value
	value  *Value    // the current value
}

func (v *Value) find(s findState) (findState, error) {
	// An empty pointer denotes the value itself.
	s.value = v
	if s.pointer[s.offset:] == "" {
		return s, nil
	}
	comp, ok := v.Value.(composite)
	if !ok {
		return s, fmt.Errorf("invalid pointer: cannot index into literal at %v", s.pointer[:s.offset])
	}

	// There must be one or more fragments.
	s.parent, s.idx, s.name = nil, 0, ""
	if !strings.HasPrefix(s.pointer[s.offset:], "/") {
		return s, fmt.Errorf("invalid pointer: lacks a forward slash prefix")
	}
	n := len("/")
	if i := strings.IndexByte(s.pointer[s.offset+n:], '/'); i >= 0 {
		n += i
	} else {
		n = len(s.pointer) - s.offset
	}
	s.offset += n

	// Unescape the name if necessary (section 4).
	name := s.pointer[s.offset-n : s.offset]
	if strings.IndexByte(name, '~') >= 0 {
		name = strings.ReplaceAll(name, "~1", "/")
		name = strings.ReplaceAll(name, "~0", "~")
	}
	name = name[len("/"):]

	// Index into the object or array.
	s.parent, s.name, s.idx = comp, name, comp.length()
	switch comp := v.Value.(type) {
	case *Object:
		for i, m := range comp.Members {
			if m.Name.Value.(Literal).equalString(name) {
				s.idx = i
				return comp.Members[i].Value.find(s)
			}
		}
	case *Array:
		if name == "-" {
			return s, errNotFound
		}
		i, err := strconv.ParseUint(name, 10, 0)
		if err != nil || (i == 0 && name != "0") {
			return s, fmt.Errorf("invalid array index: %s", name)
		}
		if i < uint64(len(comp.Elements)) {
			s.idx = int(i)
			return comp.Elements[i].find(s)
		}
	}
	return s, errNotFound
}

func (b Literal) equalString(s string) bool {
	// Fast-path: Assume there are no escape characters.
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' && bytes.IndexByte(b, '\\') < 0 {
		return string(b[len(`"`):len(b)-len(`"`)]) == s
	}
	// Slow-path: Unescape the string and then compare it.
	// TODO(dsnet): Implement allocation-free comparison.
	var s2 string
	return json.Unmarshal(b, &s2) == nil && s == s2
}
