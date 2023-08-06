// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hujson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// TODO(dsnet): Insert/remove operations on an array has O(n) complexity
// where n is the length of the array. We could improve this with more clever
// data structure that has efficient insertion, deletion, and indexing.
// One possibility is an "order statistic tree", which provide O(log n)
// behavior for the necessary operations.
// See https://en.wikipedia.org/wiki/Order_statistic_tree.

// TODO(dsnet): Name lookup on an object has O(n) complexity performing
// a linear search through all names. This can be alleviated by building
// a map of names to indexes for relevant objects. Currently, we always insert
// a new member at the end of the members list, so that operation carries an
// amortized cost of O(1).

// TODO(dsnet): Cache intermediate lookups when resolving a JSON pointer.
// Patch operations tend to operate on paths that are related.
// Caching can reduce pointer lookup from O(n) to be closer to O(1)
// where n is the number of path segments in the JSON pointer.

// TODO(dsnet): Batch sequential insert/remove operations performed
// on the same object or array. This handles the possibly common case of batch
// inserting or removing a number of consecutive members/elements.
// Pointer caching may make this optimization unnecessary.

// Patch patches the value according to the provided patch file (per RFC 6902).
// The patch file may be in the HuJSON format where comments around and within
// a value being inserted are preserved. If the patch fails to fully apply,
// the receiver value will be left in a partially mutated state.
// Use Clone to preserve the original value.
//
// It does not format the value. It is recommended that Format be called after
// applying a patch.
func (v *Value) Patch(patch []byte) error {
	ops, err := parsePatch(patch)
	if err != nil {
		return err
	}
	for i, op := range ops {
		var err error
		switch op.op {
		case "add":
			err = v.patchAdd(i, op)
		case "remove", "replace":
			err = v.patchRemoveOrReplace(i, op)
		case "move", "copy":
			err = v.patchMoveOrCopy(i, op)
		case "test":
			err = v.patchTest(i, op)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type patchOperation struct {
	op    string // "add" | "remove" | "replace" | "move" | "copy" | "test"
	path  string // used by all operations
	from  string // used by "move" and "copy"
	value Value  // used by "add", "replace", and "test"
}

func parsePatch(patch []byte) ([]patchOperation, error) {
	v, err := Parse(patch)
	if err != nil {
		return nil, err
	}
	arr, ok := v.Value.(*Array)
	if !ok {
		return nil, fmt.Errorf("hujson: patch must be a JSON array")
	}
	var ops []patchOperation
	for i, e := range arr.Elements {
		obj, ok := e.Value.(*Object)
		if !ok {
			return nil, fmt.Errorf("hujson: patch operation %d: must be a JSON object", i)
		}
		seen := make(map[string]bool)
		var op patchOperation
		for j, m := range obj.Members {
			name := m.Name.Value.(Literal).String()
			if seen[name] {
				return nil, fmt.Errorf("hujson: patch operation %d: duplicate name %q", i, m.Name.Value)
			}
			seen[name] = true
			switch name {
			case "op":
				if m.Value.Value.Kind() != '"' {
					return nil, fmt.Errorf("hujson: patch operation %d: member %q must be a JSON string", i, name)
				}
				switch opType := m.Value.Value.(Literal).String(); opType {
				case "add", "remove", "replace", "move", "copy", "test":
					op.op = opType
				default:
					return nil, fmt.Errorf("hujson: patch operation %d: unknown operation %q", i, m.Value.Value)
				}
			case "path":
				if m.Value.Value.Kind() != '"' {
					return nil, fmt.Errorf("hujson: patch operation %d: member %q must be a JSON string", i, name)
				}
				op.path = m.Value.Value.(Literal).String()
			case "from":
				if m.Value.Value.Kind() != '"' {
					return nil, fmt.Errorf("hujson: patch operation %d: member %q must be a JSON string", i, name)
				}
				op.from = m.Value.Value.(Literal).String()
			case "value":
				m.Value.BeforeExtra = obj.beforeExtraAt(j + 0).extractLeadingComments(true)
				m.Value.AfterExtra = obj.beforeExtraAt(j + 1).extractTrailingcomments(true)
				op.value = m.Value
			}
		}
		switch {
		case !seen["op"]:
			return nil, fmt.Errorf("hujson: patch operation %d: missing required member %q", i, "op")
		case !seen["path"]:
			return nil, fmt.Errorf("hujson: patch operation %d: missing required member %q", i, "path")
		case !seen["from"] && (op.op == "move" || op.op == "copy"):
			return nil, fmt.Errorf("hujson: patch operation %d: missing required member %q", i, "from")
		case !seen["value"] && (op.op == "add" || op.op == "replace" || op.op == "test"):
			return nil, fmt.Errorf("hujson: patch operation %d: missing required member %q", i, "value")
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func (v *Value) patchAdd(i int, op patchOperation) error {
	s, err := v.find(findState{pointer: op.path})
	if err != nil && (err != errNotFound || len(s.pointer) != s.offset) {
		return fmt.Errorf("hujson: patch operation %d: %v", i, err)
	}
	if s.parent == nil {
		*v = op.value // only occurs for root
	} else {
		switch comp := s.parent.(type) {
		case *Object:
			if s.idx < comp.length() {
				replaceAt(comp, s.idx, op.value)
			} else {
				insertAt(comp, s.idx, op.value)
				comp.Members[s.idx].Name.Value = String(s.name)
			}
		case *Array:
			insertAt(comp, s.idx, op.value)
		}
	}
	return nil
}

func (v *Value) patchRemoveOrReplace(i int, op patchOperation) error {
	s, err := v.find(findState{pointer: op.path})
	if err != nil {
		return fmt.Errorf("hujson: patch operation %d: %v", i, err)
	}
	if s.parent == nil {
		return fmt.Errorf("hujson: patch operation %d: cannot %s root value", i, op.op)
	}
	switch op.op {
	case "remove":
		removeAt(s.parent, s.idx)
	case "replace":
		replaceAt(s.parent, s.idx, op.value)
	}
	return nil
}

func (v *Value) patchMoveOrCopy(i int, op patchOperation) error {
	if op.from == "" || (op.op == "move" && hasPathPrefix(op.path, op.from)) {
		return fmt.Errorf("hujson: patch operation %d: cannot %s %q into %q", i, op.op, op.from, op.path)
	}
	sFrom, err := v.find(findState{pointer: op.from})
	if err != nil {
		return fmt.Errorf("hujson: patch operation %d: %v", i, err)
	}
	// TODO(dsnet): For a move operation within the same object,
	// we should simplify this as just a rename or replace.
	switch op.op {
	case "move":
		op.value = removeAt(sFrom.parent, sFrom.idx)
	case "copy":
		op.value = copyAt(sFrom.parent, sFrom.idx)
	}
	return v.patchAdd(i, op)
}

func (v *Value) patchTest(i int, op patchOperation) error {
	s, err := v.find(findState{pointer: op.path})
	if err != nil {
		return fmt.Errorf("hujson: patch operation %d: %v", i, err)
	}
	if !equalValue(*s.value, op.value) {
		return fmt.Errorf("hujson: patch operation %d: values differ at %q", i, op.path)

	}
	return nil
}

// hasPathPrefix is a stricter version of strings.HasPrefix where
// the prefix must end on a path segment boundary.
func hasPathPrefix(s, prefix string) bool {
	if strings.HasPrefix(s, prefix) {
		return len(s) == len(prefix) || s[len(prefix)] == '/'
	}
	return false
}

func equalValue(x, y Value) bool {
	// TODO(dsnet): This definition of equality is both naive and slow.
	//	* It fails to properly compare strings with invalid UTF-8.
	//	* It fails to precisely compare integers beyond ±2⁵³.
	//	* It cannot handle values greater than ±math.MaxFloat64.
	//	* Comparison of objects with duplicate names has undefined behavior.
	unmarshal := func(v Value) (vi interface{}) {
		v = v.Clone()
		v.Standardize()
		if json.Unmarshal(v.Pack(), &vi) != nil {
			return nil
		}
		return vi
	}
	vx := unmarshal(x)
	vy := unmarshal(y)
	return reflect.DeepEqual(vx, vy) && vx != nil && vy != nil
}

func (obj *Object) getAt(i int) ValueTrimmed {
	return obj.Members[i].Value.Value
}
func (obj *Object) setAt(i int, v ValueTrimmed) {
	obj.Members[i].Value.Value = v
}
func (obj *Object) insertAt(i int, v ValueTrimmed) {
	// TODO(dsnet): Use slices.Insert. See https://golang.org/issue/45955.
	obj.Members = append(obj.Members, ObjectMember{})
	copy(obj.Members[i+1:], obj.Members[i:])
	obj.Members[i] = ObjectMember{Value: Value{Value: v}}
}
func (obj *Object) removeAt(i int) ValueTrimmed {
	// TODO(dsnet): Use slices.Delete. See https://golang.org/issue/45955.
	v := obj.Members[i].Value.Value
	copy(obj.Members[i:], obj.Members[i+1:])
	obj.Members = obj.Members[:obj.length()-1]
	return v
}

func (arr *Array) getAt(i int) ValueTrimmed {
	return arr.Elements[i].Value
}
func (arr *Array) setAt(i int, v ValueTrimmed) {
	arr.Elements[i].Value = v
}
func (arr *Array) insertAt(i int, v ValueTrimmed) {
	// TODO(dsnet): Use slices.Insert. See https://golang.org/issue/45955.
	arr.Elements = append(arr.Elements, ArrayElement{})
	copy(arr.Elements[i+1:], arr.Elements[i:])
	arr.Elements[i] = ArrayElement{Value: v}
}
func (arr *Array) removeAt(i int) ValueTrimmed {
	// TODO(dsnet): Use slices.Delete. See https://golang.org/issue/45955.
	v := arr.Elements[i].Value
	copy(arr.Elements[i:], arr.Elements[i+1:])
	arr.Elements = arr.Elements[:arr.length()-1]
	return v
}

// Preserving and moving comments is impossible to perform reasonably in all
// conceivable situations given that the placement of comments is more
// a matter of human taste than it is a matter of mathematical rigor.
//
// We assume that:
//	* comments do not appear between the object member name and the colon
//	  (i.e., ObjectMember.Name.AfterExtra is nil),
//	* comments do not appear between the colon and the object member value
//	  (i.e., ObjectMember.Value.BeforeExtra is nil), and
//	* comments do not appear between the value and the comma
//	  (i.e., ObjectMember.Value.AfterExtra and ArrayElement.AfterExtra are nil).
// Such comments will be lost when patching.
//
// We further assume that:
//	* comments before an object member name and before an array element value
//	  are strongly associated with that member/element, and
//	* comments immediately after an object member value and after an
//	  array element value are strongly associated with that member/element.
// Such comments will be moved along with the member/element.
//
// Consider the following example:
//	{
//		...
//		// Comment1
//
//		// Comment2
//		"name": "value", // Comment3
//		// Comment4
//
//		// Comment5
//		...
//	}
//
// Moving "/name" will move only Comment2, Comment3, and Comment4.
// All other comments will be left alone.
//
// The above approach may perform contrary to expectation in this example:
//	{
//		// Comment1
//		"name1": "value1",
//		"name2": "value2",
//		"name3": "value3",
//	}
//
// Moving "/name1" will move Comment1. It is unclear whether Comment1 is
// strongly associated with just "name1" or the entire sequence of members
// from "name1" to "name2".

func copyAt(comp composite, i int) (v Value) {
	v.BeforeExtra = comp.beforeExtraAt(i + 0).extractLeadingComments(true)
	v.AfterExtra = comp.beforeExtraAt(i + 1).extractTrailingcomments(true)
	v.Value = comp.getAt(i).clone()
	return v
}
func replaceAt(comp composite, i int, v Value) {
	comp.beforeExtraAt(i + 0).injectLeadingComments(v.BeforeExtra)
	comp.beforeExtraAt(i + 1).injectTrailingComments(v.AfterExtra)
	comp.setAt(i, v.Value)
}
func insertAt(comp composite, i int, v Value) {
	comp.insertAt(i, v.Value)
	trailing := comp.beforeExtraAt(i + 1).extractTrailingcomments(false)
	comp.beforeExtraAt(i + 0).injectTrailingComments(trailing)
	comp.beforeExtraAt(i + 0).injectLeadingComments(v.BeforeExtra)
	comp.beforeExtraAt(i + 1).injectTrailingComments(v.AfterExtra)
}
func removeAt(comp composite, i int) (v Value) {
	v.BeforeExtra = comp.beforeExtraAt(i + 0).extractLeadingComments(false)
	v.AfterExtra = comp.beforeExtraAt(i + 1).extractTrailingcomments(false)
	if trailing := *comp.beforeExtraAt(i + 0); trailing.hasComment() {
		leading := *comp.beforeExtraAt(i + 1)
		leading = leading[consumeWhitespace(leading):]
		*comp.beforeExtraAt(i + 1) = append(trailing, leading...)
	}
	v.Value = comp.removeAt(i)
	return v
}

// injectLeadingComments injects leading comments into the bottom of b.
func (b *Extra) injectLeadingComments(leading Extra) {
	if len(leading) > 0 {
		_, currStart := b.classifyComments()
		blankLen := consumeWhitespace((*b)[currStart:])
		*b = (*b)[:currStart+blankLen]
		leading = leading[consumeWhitespace(leading):]
		if len(leading) > 0 {
			if i := bytes.LastIndexByte(*b, '\n'); i < 0 || (*b)[i:].hasComment() {
				*b = append(*b, newline...)
			}
			*b = append(*b, leading...)
		}
	}
}

// extractLeadingComments extracts leading comments from the bottom of b.
// If readonly, then the source is not mutated.
func (b *Extra) extractLeadingComments(readonly bool) (leading Extra) {
	_, currStart := b.classifyComments()
	blankLen := consumeWhitespace((*b)[currStart:])
	leading = copyBytes((*b)[currStart+blankLen:])
	if !readonly {
		*b = (*b)[:currStart+blankLen]
	}
	return leading
}

// injectTrailingComments injects trailing comments into the top of b.
func (b *Extra) injectTrailingComments(trailing Extra) {
	if len(trailing) > 0 {
		prevEnd, _ := b.classifyComments()
		if bytes.HasSuffix((*b)[:prevEnd], newline) {
			prevEnd-- // preserve trailing newline
		}
		*b = (*b)[prevEnd:]
		if trailing.hasComment() {
			if bytes.HasSuffix(trailing, newline) && bytes.HasPrefix(*b, newline) {
				trailing = trailing[:len(trailing)-1] // drop trailing newline
			}
			*b = append(copyBytes(trailing), *b...)
		}
	}
}

// extractTrailingcomments extracts trailing comments from the top of b.
// If readonly, then the source is not mutated.
func (b *Extra) extractTrailingcomments(readonly bool) (trailing Extra) {
	prevEnd, _ := b.classifyComments()
	trailing = copyBytes((*b)[:prevEnd])
	if !readonly {
		if bytes.HasSuffix(trailing, newline) {
			prevEnd-- // preserve trailing newline
		}
		*b = (*b)[prevEnd:]
	}
	return trailing
}

// classifyComments classifies comments as belonging to the previous element
// or belonging to the current element such that:
//	* b[:prevEnd] belongs to the previous element, and
//	* b[currStart:] belongs to the current element.
//
// Invariant: prevEnd <= currStart
func (b Extra) classifyComments() (prevEnd, currStart int) {
	// Scan for dividers between comment blocks.
	var firstDivider, lastDivider, numDividers int
	var n, prevNewline int
	for len(b) > n {
		nw := consumeWhitespace(b[n:])
		if prevNewline+bytes.Count(b[n:][:nw], newline) >= 2 {
			if numDividers == 0 {
				firstDivider = n
			}
			lastDivider = n
			numDividers++
		}
		n += nw

		nc := consumeComment(b[n:])
		if nc <= 0 {
			break
		}
		prevNewline = 0
		if bytes.HasSuffix(b[n:][:nc], newline) {
			prevNewline = 1 // adjust newline accounting for next iteration
		}
		n += nc
	}

	// Without dividers, a line comment starting on the first line belongs
	// to the previous element.
	if numDividers == 0 {
		nw := consumeWhitespace(b)
		nc := consumeComment(b[nw:])
		if n = nw + nc; bytes.Count(b[:n], newline) == 1 && bytes.HasSuffix(b[:n], lineCommentEnd) {
			return n, n
		}
		return 0, 0
	}

	// Ownership is more clear when there is at least one divider.
	return firstDivider, lastDivider
}
