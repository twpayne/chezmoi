// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

// JSON value parser state machine.
// Just about at the limit of what is reasonable to write by hand.
// Some parts are a bit tedious, but overall it nicely factors out the
// otherwise common code from the multiple scanning functions
// in this package (Compact, Indent, checkValid, NextValue, etc).
//
// This file starts with two simple examples using the scanner
// before diving into the scanner itself.

import "strconv"

// checkValid verifies that data is valid JSON-encoded data.
// scan is passed in for use by checkValid to avoid an allocation.
func checkValid(data []byte, scan *Scanner) error {
	scan.Reset()
	for _, c := range data {
		scan.bytes++
		if scan.Step(scan, int(c)) == ScanError {
			return scan.Err
		}
	}
	if scan.EOF() == ScanError {
		return scan.Err
	}
	return nil
}

// Validate some alleged JSON.  Return nil iff the JSON is valid.
func Validate(data []byte) error {
	s := &Scanner{}
	return checkValid(data, s)
}

// NextValue splits data after the next whole JSON value,
// returning that value and the bytes that follow it as separate slices.
// scan is passed in for use by NextValue to avoid an allocation.
func NextValue(data []byte, scan *Scanner) (value, rest []byte, err error) {
	scan.Reset()
	for i, c := range data {
		v := scan.Step(scan, int(c))
		if v >= ScanEnd {
			switch v {
			case ScanError:
				return nil, nil, scan.Err
			case ScanEnd:
				return data[0:i], data[i:], nil
			}
		}
	}
	if scan.EOF() == ScanError {
		return nil, nil, scan.Err
	}
	return data, nil, nil
}

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

// A Scanner is a JSON scanning state machine.
// Callers call scan.Reset() and then pass bytes in one at a time
// by calling scan.Step(&scan, c) for each byte.
// The return value, referred to as an opcode, tells the
// caller about significant parsing events like beginning
// and ending literals, objects, and arrays, so that the
// caller can follow along if it wishes.
// The return value ScanEnd indicates that a single top-level
// JSON value has been completed, *before* the byte that
// just got passed in.  (The indication must be delayed in order
// to recognize the end of numbers: is 123 a whole value or
// the beginning of 12345e+6?).
type Scanner struct {
	// The step is a func to be called to execute the next transition.
	// Also tried using an integer constant and a single func
	// with a switch, but using the func directly was 10% faster
	// on a 64-bit Mac Mini, and it's nicer to read.
	Step func(*Scanner, int) int

	// Reached end of top-level value.
	endTop bool

	// Stack of what we're in the middle of - array values, object keys, object values.
	parseState []int

	// Error that happened, if any.
	Err error

	// 1-byte redo (see undo method)
	redo      bool
	redoCode  int
	redoState func(*Scanner, int) int

	// total bytes consumed, updated by decoder.Decode
	bytes int64
}

// These values are returned by the state transition functions
// assigned to Scanner.state and the method Scanner.EOF.
// They give details about the current state of the scan that
// callers might be interested to know about.
// It is okay to ignore the return value of any particular
// call to Scanner.state: if one call returns ScanError,
// every subsequent call will return ScanError too.
const (
	// Continue.
	ScanContinue     = iota // uninteresting byte
	ScanBeginLiteral        // end implied by next result != scanContinue
	ScanBeginObject         // begin object
	ScanObjectKey           // just finished object key (string)
	ScanObjectValue         // just finished non-last object value
	ScanEndObject           // end object (implies scanObjectValue if possible)
	ScanBeginArray          // begin array
	ScanArrayValue          // just finished array value
	ScanEndArray            // end array (implies scanArrayValue if possible)
	ScanSkipSpace           // space byte; can skip; known to be last "continue" result

	// Stop.
	ScanEnd   // top-level value ended *before* this byte; known to be first "stop" result
	ScanError // hit an error, Scanner.err.
)

// These values are stored in the parseState stack.
// They give the current state of a composite value
// being scanned.  If the parser is inside a nested value
// the parseState describes the nested state, outermost at entry 0.
const (
	parseObjectKey   = iota // parsing object key (before colon)
	parseObjectValue        // parsing object value (after colon)
	parseArrayValue         // parsing array value
)

// reset prepares the scanner for use.
// It must be called before calling s.step.
func (s *Scanner) Reset() {
	s.Step = stateBeginValue
	s.parseState = s.parseState[0:0]
	s.Err = nil
	s.redo = false
	s.endTop = false
}

// EOF tells the scanner that the end of input has been reached.
// It returns a scan status just as s.step does.
func (s *Scanner) EOF() int {
	if s.Err != nil {
		return ScanError
	}
	if s.endTop {
		return ScanEnd
	}
	s.Step(s, ' ')
	if s.endTop {
		return ScanEnd
	}
	if s.Err == nil {
		s.Err = &SyntaxError{"unexpected end of JSON input", s.bytes}
	}
	return ScanError
}

// pushParseState pushes a new parse state p onto the parse stack.
func (s *Scanner) pushParseState(p int) {
	s.parseState = append(s.parseState, p)
}

// popParseState pops a parse state (already obtained) off the stack
// and updates s.step accordingly.
func (s *Scanner) popParseState() {
	n := len(s.parseState) - 1
	s.parseState = s.parseState[0:n]
	s.redo = false
	if n == 0 {
		s.Step = stateEndTop
		s.endTop = true
	} else {
		s.Step = stateEndValue
	}
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

// stateBeginValueOrEmpty is the state after reading `[`.
func stateBeginValueOrEmpty(s *Scanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return ScanSkipSpace
	}
	if c == ']' {
		return stateEndValue(s, c)
	}
	return stateBeginValue(s, c)
}

// stateBeginValue is the state at the beginning of the input.
func stateBeginValue(s *Scanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return ScanSkipSpace
	}
	switch c {
	case '{':
		s.Step = stateBeginStringOrEmpty
		s.pushParseState(parseObjectKey)
		return ScanBeginObject
	case '[':
		s.Step = stateBeginValueOrEmpty
		s.pushParseState(parseArrayValue)
		return ScanBeginArray
	case '"':
		s.Step = stateInString
		return ScanBeginLiteral
	case '-':
		s.Step = stateNeg
		return ScanBeginLiteral
	case '0': // beginning of 0.123
		s.Step = state0
		return ScanBeginLiteral
	case 't': // beginning of true
		s.Step = stateT
		return ScanBeginLiteral
	case 'f': // beginning of false
		s.Step = stateF
		return ScanBeginLiteral
	case 'n': // beginning of null
		s.Step = stateN
		return ScanBeginLiteral
	}
	if '1' <= c && c <= '9' { // beginning of 1234.5
		s.Step = state1
		return ScanBeginLiteral
	}
	return s.error(c, "looking for beginning of value")
}

// stateBeginStringOrEmpty is the state after reading `{`.
func stateBeginStringOrEmpty(s *Scanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return ScanSkipSpace
	}
	if c == '}' {
		n := len(s.parseState)
		s.parseState[n-1] = parseObjectValue
		return stateEndValue(s, c)
	}
	return stateBeginString(s, c)
}

// stateBeginString is the state after reading `{"key": value,`.
func stateBeginString(s *Scanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return ScanSkipSpace
	}
	if c == '"' {
		s.Step = stateInString
		return ScanBeginLiteral
	}
	return s.error(c, "looking for beginning of object key string")
}

// stateEndValue is the state after completing a value,
// such as after reading `{}` or `true` or `["x"`.
func stateEndValue(s *Scanner, c int) int {
	n := len(s.parseState)
	if n == 0 {
		// Completed top-level before the current byte.
		s.Step = stateEndTop
		s.endTop = true
		return stateEndTop(s, c)
	}
	if c <= ' ' && isSpace(rune(c)) {
		s.Step = stateEndValue
		return ScanSkipSpace
	}
	ps := s.parseState[n-1]
	switch ps {
	case parseObjectKey:
		if c == ':' {
			s.parseState[n-1] = parseObjectValue
			s.Step = stateBeginValue
			return ScanObjectKey
		}
		return s.error(c, "after object key")
	case parseObjectValue:
		if c == ',' {
			s.parseState[n-1] = parseObjectKey
			s.Step = stateBeginString
			return ScanObjectValue
		}
		if c == '}' {
			s.popParseState()
			return ScanEndObject
		}
		return s.error(c, "after object key:value pair")
	case parseArrayValue:
		if c == ',' {
			s.Step = stateBeginValue
			return ScanArrayValue
		}
		if c == ']' {
			s.popParseState()
			return ScanEndArray
		}
		return s.error(c, "after array element")
	}
	return s.error(c, "")
}

// stateEndTop is the state after finishing the top-level value,
// such as after reading `{}` or `[1,2,3]`.
// Only space characters should be seen now.
func stateEndTop(s *Scanner, c int) int {
	if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
		// Complain about non-space byte on next call.
		s.error(c, "after top-level value")
	}
	return ScanEnd
}

// stateInString is the state after reading `"`.
func stateInString(s *Scanner, c int) int {
	if c == '"' {
		s.Step = stateEndValue
		return ScanContinue
	}
	if c == '\\' {
		s.Step = stateInStringEsc
		return ScanContinue
	}
	if c < 0x20 {
		return s.error(c, "in string literal")
	}
	return ScanContinue
}

// stateInStringEsc is the state after reading `"\` during a quoted string.
func stateInStringEsc(s *Scanner, c int) int {
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		s.Step = stateInString
		return ScanContinue
	}
	if c == 'u' {
		s.Step = stateInStringEscU
		return ScanContinue
	}
	return s.error(c, "in string escape code")
}

// stateInStringEscU is the state after reading `"\u` during a quoted string.
func stateInStringEscU(s *Scanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.Step = stateInStringEscU1
		return ScanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU1 is the state after reading `"\u1` during a quoted string.
func stateInStringEscU1(s *Scanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.Step = stateInStringEscU12
		return ScanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU12 is the state after reading `"\u12` during a quoted string.
func stateInStringEscU12(s *Scanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.Step = stateInStringEscU123
		return ScanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU123 is the state after reading `"\u123` during a quoted string.
func stateInStringEscU123(s *Scanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.Step = stateInString
		return ScanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateNeg is the state after reading `-` during a number.
func stateNeg(s *Scanner, c int) int {
	if c == '0' {
		s.Step = state0
		return ScanContinue
	}
	if '1' <= c && c <= '9' {
		s.Step = state1
		return ScanContinue
	}
	return s.error(c, "in numeric literal")
}

// state1 is the state after reading a non-zero integer during a number,
// such as after reading `1` or `100` but not `0`.
func state1(s *Scanner, c int) int {
	if '0' <= c && c <= '9' {
		s.Step = state1
		return ScanContinue
	}
	return state0(s, c)
}

// state0 is the state after reading `0` during a number.
func state0(s *Scanner, c int) int {
	if c == '.' {
		s.Step = stateDot
		return ScanContinue
	}
	if c == 'e' || c == 'E' {
		s.Step = stateE
		return ScanContinue
	}
	return stateEndValue(s, c)
}

// stateDot is the state after reading the integer and decimal point in a number,
// such as after reading `1.`.
func stateDot(s *Scanner, c int) int {
	if '0' <= c && c <= '9' {
		s.Step = stateDot0
		return ScanContinue
	}
	return s.error(c, "after decimal point in numeric literal")
}

// stateDot0 is the state after reading the integer, decimal point, and subsequent
// digits of a number, such as after reading `3.14`.
func stateDot0(s *Scanner, c int) int {
	if '0' <= c && c <= '9' {
		s.Step = stateDot0
		return ScanContinue
	}
	if c == 'e' || c == 'E' {
		s.Step = stateE
		return ScanContinue
	}
	return stateEndValue(s, c)
}

// stateE is the state after reading the mantissa and e in a number,
// such as after reading `314e` or `0.314e`.
func stateE(s *Scanner, c int) int {
	if c == '+' {
		s.Step = stateESign
		return ScanContinue
	}
	if c == '-' {
		s.Step = stateESign
		return ScanContinue
	}
	return stateESign(s, c)
}

// stateESign is the state after reading the mantissa, e, and sign in a number,
// such as after reading `314e-` or `0.314e+`.
func stateESign(s *Scanner, c int) int {
	if '0' <= c && c <= '9' {
		s.Step = stateE0
		return ScanContinue
	}
	return s.error(c, "in exponent of numeric literal")
}

// stateE0 is the state after reading the mantissa, e, optional sign,
// and at least one digit of the exponent in a number,
// such as after reading `314e-2` or `0.314e+1` or `3.14e0`.
func stateE0(s *Scanner, c int) int {
	if '0' <= c && c <= '9' {
		s.Step = stateE0
		return ScanContinue
	}
	return stateEndValue(s, c)
}

// stateT is the state after reading `t`.
func stateT(s *Scanner, c int) int {
	if c == 'r' {
		s.Step = stateTr
		return ScanContinue
	}
	return s.error(c, "in literal true (expecting 'r')")
}

// stateTr is the state after reading `tr`.
func stateTr(s *Scanner, c int) int {
	if c == 'u' {
		s.Step = stateTru
		return ScanContinue
	}
	return s.error(c, "in literal true (expecting 'u')")
}

// stateTru is the state after reading `tru`.
func stateTru(s *Scanner, c int) int {
	if c == 'e' {
		s.Step = stateEndValue
		return ScanContinue
	}
	return s.error(c, "in literal true (expecting 'e')")
}

// stateF is the state after reading `f`.
func stateF(s *Scanner, c int) int {
	if c == 'a' {
		s.Step = stateFa
		return ScanContinue
	}
	return s.error(c, "in literal false (expecting 'a')")
}

// stateFa is the state after reading `fa`.
func stateFa(s *Scanner, c int) int {
	if c == 'l' {
		s.Step = stateFal
		return ScanContinue
	}
	return s.error(c, "in literal false (expecting 'l')")
}

// stateFal is the state after reading `fal`.
func stateFal(s *Scanner, c int) int {
	if c == 's' {
		s.Step = stateFals
		return ScanContinue
	}
	return s.error(c, "in literal false (expecting 's')")
}

// stateFals is the state after reading `fals`.
func stateFals(s *Scanner, c int) int {
	if c == 'e' {
		s.Step = stateEndValue
		return ScanContinue
	}
	return s.error(c, "in literal false (expecting 'e')")
}

// stateN is the state after reading `n`.
func stateN(s *Scanner, c int) int {
	if c == 'u' {
		s.Step = stateNu
		return ScanContinue
	}
	return s.error(c, "in literal null (expecting 'u')")
}

// stateNu is the state after reading `nu`.
func stateNu(s *Scanner, c int) int {
	if c == 'l' {
		s.Step = stateNul
		return ScanContinue
	}
	return s.error(c, "in literal null (expecting 'l')")
}

// stateNul is the state after reading `nul`.
func stateNul(s *Scanner, c int) int {
	if c == 'l' {
		s.Step = stateEndValue
		return ScanContinue
	}
	return s.error(c, "in literal null (expecting 'l')")
}

// stateError is the state after reaching a syntax error,
// such as after reading `[1}` or `5.1.2`.
func stateError(s *Scanner, c int) int {
	return ScanError
}

// error records an error and switches to the error state.
func (s *Scanner) error(c int, context string) int {
	s.Step = stateError
	s.Err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, s.bytes}
	return ScanError
}

// quoteChar formats c as a quoted character literal
func quoteChar(c int) string {
	// special cases - different from quoted strings
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}

	// use quoted string with different quotation marks
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}

// undo causes the scanner to return scanCode from the next state transition.
// This gives callers a simple 1-byte undo mechanism.
func (s *Scanner) undo(scanCode int) {
	if s.redo {
		panic("json: invalid use of scanner")
	}
	s.redoCode = scanCode
	s.redoState = s.Step
	s.Step = stateRedo
	s.redo = true
}

// stateRedo helps implement the scanner's 1-byte undo.
func stateRedo(s *Scanner, c int) int {
	s.redo = false
	s.Step = s.redoState
	return s.redoCode
}
