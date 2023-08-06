// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hujson

import (
	"bytes"
	"unicode"
)

// Standardize strips any features specific to HuJSON from b,
// making it compliant with standard JSON per RFC 8259.
// All comments and trailing commas are replaced with a space character
// in order to preserve the original line numbers and byte offsets.
// If an error is encountered, then b is returned as is along with the error.
func Standardize(b []byte) ([]byte, error) {
	ast, err := Parse(b)
	if err != nil {
		return b, err
	}
	ast.Standardize()
	return ast.Pack(), nil
}

// Minimize removes all whitespace, comments, and trailing commas from b,
// making it compliant with standard JSON per RFC 8259.
// If an error is encountered, then b is returned as is along with the error.
func Minimize(b []byte) ([]byte, error) {
	ast, err := Parse(b)
	if err != nil {
		return b, err
	}
	ast.Minimize()
	return ast.Pack(), nil
}

// Format formats b according to some opinionated heuristics for
// how HuJSON should look. The exact output may change over time.
// It is the equivalent of `go fmt` but for HuJSON.
//
// If the input is standard JSON, then the output will remain standard.
// Format is idempotent such that formatting already formatted HuJSON
// results in no changes.
// If an error is encountered, then b is returned as is along with the error.
func Format(b []byte) ([]byte, error) {
	ast, err := Parse(b)
	if err != nil {
		return b, err
	}
	ast.Format()
	return ast.Pack(), nil
}

const punchCardWidth = 80

var (
	newline        = []byte("\n")
	twoNewlines    = []byte("\n\n")
	endlineWindows = []byte("\r\n")
	endlineMacOSX  = []byte("\n\r")
	carriageReturn = []byte("\r")
	space          = []byte(" ")
)

// Format formats the value according to some opinionated heuristics for
// how HuJSON should look. The exact output may change over time.
// It is the equivalent of `go fmt` but for HuJSON.
//
// If the input is standard JSON, then the output will remain standard.
// Format is idempotent such that formatting already formatted HuJSON
// results in no changes.
func (v *Value) Format() {
	// Format leading extra.
	v.BeforeExtra.format(0, formatOptions{})
	v.BeforeExtra = v.BeforeExtra[consumeWhitespace(v.BeforeExtra):] // never has leading whitespace
	// Format the value.
	needExpand := make(map[composite]bool)
	isStandard := v.IsStandard()
	v.normalize()
	v.expandComposites(needExpand)
	v.formatWhitespace(0, needExpand, isStandard)
	v.alignObjectValues()
	// Format trailing extra.
	v.AfterExtra.format(0, formatOptions{})
	v.AfterExtra = append(bytes.TrimRightFunc(v.AfterExtra, unicode.IsSpace), '\n') // always has exactly one trailing newline

	v.UpdateOffsets()
}

// Range iterates through a Value in depth-first order and
// calls f for each value (including the root value).
// It stops iteration when f returns false.
func (v *Value) Range(f func(v *Value) bool) bool {
	if !f(v) {
		return false
	}
	if comp, ok := v.Value.(composite); ok {
		return comp.rangeValues(func(v2 *Value) bool {
			return v2.Range(f)
		})
	}
	return true
}

// normalize performs simple normalization changes. In particular, it:
//   - normalizes strings,
//   - normalizes empty objects and arrays as simply {} or [],
//   - normalizes whitespace between names and colons,
//   - normalizes whitespace between values and commas.
//
// It always returns true to be compatible with composite.rangeValues.
func (v *Value) normalize() bool {
	switch v2 := v.Value.(type) {
	case Literal:
		// Normalize string if there are escape characters.
		if v2.Kind() == '"' && bytes.IndexByte(v2, '\\') >= 0 {
			v.Value = String(v2.String())
		}
	case composite:
		// Cleanup for empty objects and arrays.
		if v2.length() == 0 {
			// If there is only whitespace, then remove the whitespace.
			if !v2.afterExtra().hasComment() {
				*v2.afterExtra() = nil
			}
			break
		}

		// If there is only whitespace between the name and colon,
		// or between the value and comma, then remove the whitespace.
		v2.rangeValues(func(v *Value) bool {
			if !v.AfterExtra.hasComment() {
				v.AfterExtra = nil
			}
			return true
		})

		// Normalize all sub-values.
		v2.rangeValues((*Value).normalize)
	}
	return true
}

// lineStats carries statistics about a sequence of lines.
type lineStats struct {
	firstLength int
	lastLength  int
	multiline   bool // false implies firstLength == lastLength
}

// expandComposites populates needExpand with the set of composite values
// that need to be expanded (i.e., print each member/element on a new line).
// This method is pure and does not mutate the AST.
func (v *Value) expandComposites(needExpand map[composite]bool) (stats lineStats) {
	switch v2 := v.Value.(type) {
	case Literal:
		stats = lineStats{len(v2), len(v2), false}
	case composite:
		// Every object or array is either fully inlined or fully expanded.
		// This simplifies machine-modification of HuJSON so that the mutation
		// can easily determine which mode it is currently in.
		//
		// If any whitespace after a '{', '[', or ',' or before a '}' or ']'
		// contains a newline, then we always expand the object or array.
		var expand bool

		// Keep track of line lengths.
		var lineLength int
		var lineLengths []int
		updateStats := func(s lineStats) {
			lineLength += s.firstLength
			if s.multiline {
				lineLengths = append(lineLengths, lineLength)
				lineLength = s.lastLength
			}
		}

		// Iterate through all members/elements in an object/array.
		switch v2 := v2.(type) {
		case *Object:
			lineLength += len("{")
			for i := range v2.Members {
				name := &v2.Members[i].Name
				value := &v2.Members[i].Value
				expand = expand || name.BeforeExtra.hasNewline()
				updateStats(name.BeforeExtra.lineStats())
				updateStats(name.expandComposites(needExpand))
				updateStats(name.AfterExtra.lineStats())
				lineLength += len(": ")
				updateStats(value.BeforeExtra.lineStats())
				updateStats(value.expandComposites(needExpand))
				updateStats(value.AfterExtra.lineStats())
				lineLength += len(", ")
			}
			lineLength += len("}")

			// Always expand multiline objects with more than 1 member.
			expand = expand || v2.length() > 1 && stats.multiline
		case *Array:
			lineLength += len("[")
			for i := range v2.Elements {
				value := &v2.Elements[i]
				expand = expand || value.BeforeExtra.hasNewline()
				updateStats(value.BeforeExtra.lineStats())
				updateStats(value.expandComposites(needExpand))
				updateStats(value.AfterExtra.lineStats())
				lineLength += len(", ")
			}
			lineLength += len("]")
		}
		if last := v2.lastValue(); last != nil {
			expand = expand || last.AfterExtra.hasNewline()
		}
		expand = expand || v2.afterExtra().hasNewline()

		// Update the block statistics.
		lineLengths = append(lineLengths, lineLength)
		stats = lineStats{
			firstLength: lineLengths[0],
			lastLength:  lineLengths[len(lineLengths)-1],
			multiline:   len(lineLengths) > 1,
		}
		for i := 0; !expand && i < len(lineLengths); i++ {
			expand = lineLengths[i] > punchCardWidth
		}

		if expand {
			stats = lineStats{len("{"), len("}"), true}
			stats.firstLength += v2.beforeExtraAt(0).lineStats().firstLength
			needExpand[v2] = expand
		}
	}
	return stats
}

func (b Extra) lineStats() (stats lineStats) {
	// length is the approximate length of the comments.
	length := func(b []byte) (n int) {
		for {
			b = b[consumeWhitespace(b):]
			switch {
			case bytes.HasPrefix(b, lineCommentStart):
				return n + len(" ") + len(b) // line comment must go to the end
			case bytes.HasPrefix(b, blockCommentStart):
				nc := consumeComment(b)
				if nc <= 0 {
					return n + len(" ") + len(b) // truncated block comment must go to the end
				}
				n += len(" ") + nc
				b = b[nc:]
				continue
			default:
				if n > 0 {
					n += len(" ") // account for padding space after block comment
				}
				return n
			}
		}
	}
	if !bytes.Contains(b, newline) {
		n := length(b)
		return lineStats{n, n, false}
	} else {
		first := b[:bytes.IndexByte(b, '\n')]
		last := b[bytes.LastIndexByte(b, '\n')+len("\n"):]
		return lineStats{length(first), length(last), true}
	}
}

// formatWhitespace mutates the AST and formats whitespace to ensure
// consistent indentation and expansion of objects and arrays.
func (v *Value) formatWhitespace(depth int, needExpand map[composite]bool, standardize bool) {
	if comp, ok := v.Value.(composite); ok {
		expand := needExpand[comp]

		// Format all members/elements in an object/array.
		switch comp := comp.(type) {
		case *Object:
			for i := range comp.Members {
				name := &comp.Members[i].Name
				value := &comp.Members[i].Value

				// Format extra before name.
				name.BeforeExtra.format(depth+1, formatOptions{
					ensureLeadingNewline:    expand,
					removeLeadingEmptyLines: i == 0,
					appendSpaceIfEmpty:      i != 0,
				})
				// Format the name.
				name.formatWhitespace(depth+1, needExpand, standardize)
				// Format extra after name and before colon.
				name.AfterExtra.format(depth+2, formatOptions{
					removeLeadingEmptyLines:  true,
					removeTrailingEmptyLines: true,
				})
				// Format extra after colon and before value.
				value.BeforeExtra.format(depth+2, formatOptions{
					removeLeadingEmptyLines:  true,
					removeTrailingEmptyLines: true,
					appendSpaceIfEmpty:       true,
				})
				// Format the value.
				depthOffset := 0
				if expand {
					depthOffset++
				}
				if name.AfterExtra.hasNewline() || value.BeforeExtra.hasNewline() {
					depthOffset++
				}
				value.formatWhitespace(depth+depthOffset, needExpand, standardize)
				// Format extra after value and before comma.
				value.AfterExtra.format(depth+2, formatOptions{
					removeLeadingEmptyLines:  true,
					removeTrailingEmptyLines: true,
				})
			}
		case *Array:
			for i := range comp.Elements {
				value := &comp.Elements[i]

				// Format extra before value.
				value.BeforeExtra.format(depth+1, formatOptions{
					ensureLeadingNewline:    expand,
					removeLeadingEmptyLines: i == 0,
					appendSpaceIfEmpty:      i != 0,
				})
				// Format the value.
				depthOffset := 0
				if expand {
					depthOffset++
				}
				value.formatWhitespace(depth+depthOffset, needExpand, standardize)
				// Format extra after value and before comma.
				value.AfterExtra.format(depth+2, formatOptions{
					removeLeadingEmptyLines:  true,
					removeTrailingEmptyLines: true,
				})
			}
		}

		// Format the extra before the closing '}' or ']'.
		comp.afterExtra().format(depth+1, formatOptions{
			ensureTrailingNewline:    expand,
			removeLeadingEmptyLines:  comp.length() == 0,
			removeTrailingEmptyLines: true,
			unindentLastLine:         true,
		})

		// Normalize presence of trailing comma.
		surroundedComma := comp.lastValue() != nil && len(comp.lastValue().AfterExtra) > 0 && len(*comp.afterExtra()) > 0
		switch {
		// Avoid a trailing comma for a non-expanded object or array.
		case !expand && !surroundedComma:
			setTrailingComma(comp, false)
		// Otherwise, emit a trailing comma (unless this need to be standard).
		case expand && !standardize:
			setTrailingComma(comp, true)
		}
	}
}

type formatOptions struct {
	ensureLeadingNewline     bool
	ensureTrailingNewline    bool
	removeLeadingEmptyLines  bool
	removeTrailingEmptyLines bool
	unindentLastLine         bool
	appendSpaceIfEmpty       bool
}

func (b *Extra) format(depth int, opts formatOptions) {
	// Remove carriage returns to normalize output across operating systems.
	if bytes.IndexByte(*b, '\r') >= 0 {
		*b = bytes.ReplaceAll(*b, endlineWindows, newline)
		*b = bytes.ReplaceAll(*b, endlineMacOSX, newline)
		*b = bytes.ReplaceAll(*b, carriageReturn, space)
	}

	in := *b
	var out []byte // TODO(dsnet): Cache this in sync.Pool?

	// Inject a leading newline if not present in the input.
	if opts.ensureLeadingNewline && !in.hasNewline() {
		out = append(out, '\n')
	}

	// Iterate over every paragraph in the comment.
	for len(in) > 0 {
		// Handle whitespace.
		if n := consumeWhitespace(in); n > 0 {
			nl := bytes.Count(in[:n], newline)
			if nl > 2 {
				nl = 2 // never allow more than one blank line
			}
			for i := 0; i < nl; i++ {
				out = append(out, '\n')
			}
			in = in[n:]
			continue
		}

		// Handle comments.
		n := consumeComment(in)
		if n <= 0 {
			return // invalid comment
		}

		// Emit leading whitespace.
		if bytes.HasSuffix(out, newline) {
			out = appendIndent(out, depth)
		} else {
			out = append(out, ' ')
		}

		// Copy single-line comment to the output verbatim.
		comment := in[:n]
		if bytes.HasPrefix(comment, lineCommentStart) || !comment.hasNewline() {
			comment = bytes.TrimRightFunc(comment, unicode.IsSpace) // trim trailing whitespace
			if bytes.HasPrefix(comment, lineCommentStart) {
				n-- // leave newline for next iteration of comment
			}
			out = append(out, comment...) // single-line comments preserved verbatim
			in = in[n:]
			continue
		}

		// Format multi-line block comments and copy to the output.
		lines := bytes.Split(comment, newline) // len(lines) >= 2 since at least one '\n' exists
		var firstLine []byte                   // first non-empty line after blockCommentStart
		var hasEmptyLine bool
		for i, line := range lines {
			line = bytes.TrimRightFunc(line, unicode.IsSpace) // trim trailing whitespace
			if len(firstLine) == 0 && len(line) > 0 && i > 0 {
				firstLine = line
			}
			hasEmptyLine = hasEmptyLine || len(line) == 0
			lines[i] = line
		}

		// Compute the longest common prefix
		commonPrefix := firstLine
		for i, line := range lines[1:] {
			if len(line) == 0 {
				continue // ignore empty lines
			}

			// If the last line is just "*/" with preceding whitespace, then
			// ignore any whitespace as part of the common prefix.
			// Instead, copy the whitespace from the common prefix.
			isLast := bytes.HasSuffix(line, blockCommentEnd)
			if isLast && consumeWhitespace(line)+len(blockCommentEnd) == len(line) {
				prefixLen := consumeWhitespace(commonPrefix)
				lines[i+1] = append(commonPrefix[:prefixLen:prefixLen], blockCommentEnd...)
				break
			}

			// Check for longest common prefix.
			for i := 0; i < len(line) && i < len(commonPrefix); i++ {
				if line[i] != commonPrefix[i] {
					commonPrefix = commonPrefix[:i]
					continue
				}
			}
		}

		// Indent every line and copy to output.
		prefixLen := consumeWhitespace(commonPrefix)
		starAligned := !hasEmptyLine && len(commonPrefix) > prefixLen && commonPrefix[prefixLen] == '*'
		out = append(out, lines[0]...)
		out = append(out, '\n')
		for _, line := range lines[1:] {
			if len(line) > 0 {
				out = appendIndent(out, depth)
				if starAligned {
					out = append(out, ' ')
				}
				out = append(out, line[prefixLen:]...)
			}
			out = append(out, '\n')
		}
		out = bytes.TrimRight(out, "\n")
		in = in[n:]
	}

	// Inject a trailing newline if not present in the input.
	if opts.ensureTrailingNewline && !bytes.HasSuffix(out, newline) {
		out = append(out, '\n')
	}
	// Remove all leading empty lines.
	for opts.removeLeadingEmptyLines && bytes.HasPrefix(out, twoNewlines) {
		out = out[1:]
	}
	// Remove all trailing empty lines.
	for opts.removeTrailingEmptyLines && bytes.HasSuffix(out, twoNewlines) {
		out = out[:len(out)-1]
	}
	// If the whitespace ends on a newline, append the necessary indentation.
	// Otherwise, emit a space if we did not end on a new line.
	if bytes.HasSuffix(out, newline) {
		if opts.unindentLastLine {
			depth--
		}
		out = appendIndent(out, depth)
	} else if len(out) > 0 {
		out = append(out, ' ')
	}
	// Emit a space if the output is empty.
	if opts.appendSpaceIfEmpty && len(out) == 0 {
		out = append(out, ' ')
	}

	// Copy intermediate output to the receiver.
	if !bytes.Equal(*b, out) {
		*b = append((*b)[:0], out...)
	}
}

// alignObjectValues aligns object values by inserting spaces after the name
// so that the values are aligned to the same column.
//
// It always returns true to be compatible with composite.rangeValues.
func (v *Value) alignObjectValues() bool {
	// TODO(dsnet): This is broken for non-monospace, non-narrow characters.
	// This is hard to fix as even `go fmt` suffers from this problem.
	// See https://golang.org/issue/8273.
	if obj, ok := v.Value.(*Object); ok {
		type row struct {
			extra  *Extra // pointer to extra after colon and before value
			length int    // length from start of name to end of extra
		}
		var rows []row
		alignRows := func() {
			// TODO(dsnet): Should we break apart rows if the number of spaces
			// to insert exceeds some threshold?

			// Compute the maximum width.
			var max int
			for _, row := range rows {
				if max < row.length {
					max = row.length
				}
			}
			// Align every row up to that width.
			for _, row := range rows {
				for n := max - row.length; n > 0; n-- {
					*row.extra = append(*row.extra, ' ')
				}
			}
			// Reset the sequence of rows.
			rows = rows[:0]
		}
		var indentSuffix []byte
		for i := range obj.Members {
			name := &obj.Members[i].Name
			value := &obj.Members[i].Value

			// Whitespace right before name must have a newline and
			// everything after the name until the comma cannot have newlines.
			if !name.BeforeExtra.hasNewline() ||
				name.hasNewline(false) ||
				name.AfterExtra.hasNewline() ||
				value.BeforeExtra.hasNewline() ||
				value.hasNewline(false) ||
				value.AfterExtra.hasNewline() {
				alignRows()
				continue
			}

			// If there are multiple newlines or the indentSuffix mismatches,
			// then this is the start of a new block or rows to align.
			if bytes.Count(name.BeforeExtra, newline) > 1 || !bytes.HasSuffix(name.BeforeExtra, indentSuffix) {
				alignRows() // flush the current block or rows
			}

			rows = append(rows, row{
				extra:  &value.BeforeExtra,
				length: len(name.Value.(Literal)) + len(name.AfterExtra) + len(":") + len(value.BeforeExtra),
			})
		}
		alignRows()
	}

	// Recursively align all sub-objects.
	if comp, ok := v.Value.(composite); ok {
		comp.rangeValues((*Value).alignObjectValues)
	}
	return true
}

func (v Value) hasNewline(checkTopLevelExtra bool) bool {
	if checkTopLevelExtra && (v.BeforeExtra.hasNewline() || v.AfterExtra.hasNewline()) {
		return true
	}
	if comp, ok := v.Value.(composite); ok {
		return !comp.rangeValues(func(v *Value) bool {
			return !v.hasNewline(true)
		})
	}
	return false
}

func (b Extra) hasNewline() bool {
	return bytes.IndexByte(b, '\n') >= 0
}

func appendIndent(b []byte, n int) []byte {
	for i := 0; i < n; i++ {
		b = append(b, '\t')
	}
	return b
}
