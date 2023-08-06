// Package assert provides type-safe assertions with clean error messages.
package assert

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/repr"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
)

func objectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, eok := expected.([]byte)
	act, aok := actual.([]byte)

	if eok && aok {
		return bytes.Equal(exp, act)
	}

	return reflect.DeepEqual(expected, actual)
}

// Compare two values for equality and return true or false.
func Compare[T any](t testing.TB, x, y T) bool {
	return objectsAreEqual(x, y)
}

// Equal asserts that "expected" and "actual" are equal.
//
// If they are not, a diff of the Go representation of the values will be displayed.
func Equal[T any](t testing.TB, expected, actual T, msgAndArgs ...interface{}) {
	if objectsAreEqual(expected, actual) {
		return
	}
	t.Helper()
	msg := formatMsgAndArgs("Expected values to be equal:", msgAndArgs...)
	t.Fatalf("%s\n%s", msg, diff(expected, actual))
}

// NotEqual asserts that "expected" is not equal to "actual".
//
// If they are equal the expected value will be displayed.
func NotEqual[T any](t testing.TB, expected, actual T, msgAndArgs ...interface{}) {
	if !objectsAreEqual(expected, actual) {
		return
	}
	t.Helper()
	msg := formatMsgAndArgs("Expected values to not be equal but both were:", msgAndArgs...)
	t.Fatalf("%s\n%s", msg, repr.String(expected, repr.Indent("  ")))
}

// Contains asserts that "haystack" contains "needle".
func Contains(t testing.TB, haystack string, needle string, msgAndArgs ...interface{}) {
	if strings.Contains(haystack, needle) {
		return
	}
	t.Helper()
	msg := formatMsgAndArgs("Haystack does not contain needle.", msgAndArgs...)
	t.Fatalf("%s\nNeedle: %q\nHaystack: %q\n", msg, needle, haystack)
}

// NotContains asserts that "haystack" does not contain "needle".
func NotContains(t testing.TB, haystack string, needle string, msgAndArgs ...interface{}) {
	if !strings.Contains(haystack, needle) {
		return
	}
	t.Helper()
	msg := formatMsgAndArgs("Haystack should not contain needle.", msgAndArgs...)
	quotedHaystack, quotedNeedle, positions := needlePosition(haystack, needle)
	t.Fatalf("%s\nNeedle: %s\nHaystack: %s\n          %s\n", msg, quotedNeedle, quotedHaystack, positions)
}

// Zero asserts that a value is its zero value.
func Zero[T any](t testing.TB, value T, msgAndArgs ...interface{}) {
	var zero T
	if objectsAreEqual(value, zero) {
		return
	}
	val := reflect.ValueOf(value)
	if (val.Kind() == reflect.Slice || val.Kind() == reflect.Map || val.Kind() == reflect.Array) && val.Len() == 0 {
		return
	}
	t.Helper()
	msg := formatMsgAndArgs("Expected a zero value but got:", msgAndArgs...)
	t.Fatalf("%s\n%s", msg, repr.String(value, repr.Indent("  ")))
}

// NotZero asserts that a value is not its zero value.
func NotZero[T any](t testing.TB, value T, msgAndArgs ...interface{}) {
	var zero T
	if !objectsAreEqual(value, zero) {
		val := reflect.ValueOf(value)
		if !((val.Kind() == reflect.Slice || val.Kind() == reflect.Map || val.Kind() == reflect.Array) && val.Len() == 0) {
			return
		}
	}
	t.Helper()
	msg := formatMsgAndArgs("Did not expect the zero value:", msgAndArgs...)
	t.Fatalf("%s\n%s", msg, repr.String(value))
}

// EqualError asserts that either an error is non-nil and that its message is what is expected,
// or that error is nil if the expected message is empty.
func EqualError(t testing.TB, err error, errString string, msgAndArgs ...interface{}) {
	if err == nil && errString == "" {
		return
	}
	t.Helper()
	if err == nil {
		t.Fatal(formatMsgAndArgs("Expected an error", msgAndArgs...))
	}
	if err.Error() != errString {
		msg := formatMsgAndArgs("Error message not as expected:", msgAndArgs...)
		t.Fatalf("%s\n%s", msg, diff(errString, err.Error()))
	}
}

// IsError asserts than any error in "err"'s tree matches "target".
func IsError(t testing.TB, err, target error, msgAndArgs ...interface{}) {
	if errors.Is(err, target) {
		return
	}
	t.Helper()
	t.Fatal(formatMsgAndArgs(fmt.Sprintf("Error tree %q should contain error %q", err, target), msgAndArgs...))
}

// NotIsError asserts than no error in "err"'s tree matches "target".
func NotIsError(t testing.TB, err, target error, msgAndArgs ...interface{}) {
	if !errors.Is(err, target) {
		return
	}
	t.Helper()
	t.Fatal(formatMsgAndArgs(fmt.Sprintf("Error tree %q should NOT contain error %q", err, target), msgAndArgs...))
}

// Error asserts that an error is not nil.
func Error(t testing.TB, err error, msgAndArgs ...interface{}) {
	if err != nil {
		return
	}
	t.Helper()
	t.Fatal(formatMsgAndArgs("Expected an error", msgAndArgs...))
}

// NoError asserts that an error is nil.
func NoError(t testing.TB, err error, msgAndArgs ...interface{}) {
	if err == nil {
		return
	}
	t.Helper()
	msg := formatMsgAndArgs("Did not expect an error but got:", msgAndArgs...)
	t.Fatalf("%s\n%s", msg, err)
}

// True asserts that an expression is true.
func True(t testing.TB, ok bool, msgAndArgs ...interface{}) {
	if ok {
		return
	}
	t.Helper()
	t.Fatal(formatMsgAndArgs("Expected expression to be true", msgAndArgs...))
}

// False asserts that an expression is false.
func False(t testing.TB, ok bool, msgAndArgs ...interface{}) {
	if !ok {
		return
	}
	t.Helper()
	t.Fatal(formatMsgAndArgs("Expected expression to be false", msgAndArgs...))
}

// Panics asserts that the given function panics.
func Panics(t testing.TB, fn func(), msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if recover() == nil {
			msg := formatMsgAndArgs("Expected function to panic", msgAndArgs...)
			t.Fatal(msg)
		}
	}()
	fn()
}

// NotPanics asserts that the given function does not panic.
func NotPanics(t testing.TB, fn func(), msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if err := recover(); err != nil {
			msg := formatMsgAndArgs("Expected function not to panic", msgAndArgs...)
			t.Fatalf("%s\nPanic: %v", msg, err)
		}
	}()
	fn()
}

func diff[T any](before, after T) string {
	var lhss, rhss string
	// Special case strings so we get nice diffs.
	if l, ok := any(before).(string); ok {
		lhss = l
		rhss = any(after).(string)
	} else {
		lhss = repr.String(before, repr.Indent("  ")) + "\n"
		rhss = repr.String(after, repr.Indent("  ")) + "\n"
	}
	edits := myers.ComputeEdits("a.txt", lhss, rhss)
	lines := strings.Split(fmt.Sprint(gotextdiff.ToUnified("expected.txt", "actual.txt", lhss, edits)), "\n")
	if len(lines) < 3 {
		return ""
	}
	return strings.Join(lines[3:], "\n")
}

func formatMsgAndArgs(dflt string, msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return dflt
	}
	return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
}

func needlePosition(haystack, needle string) (quotedHaystack, quotedNeedle, positions string) {
	quotedNeedle = strconv.Quote(needle)
	quotedNeedle = quotedNeedle[1 : len(quotedNeedle)-1]
	quotedHaystack = strconv.Quote(haystack)
	rawPositions := strings.ReplaceAll(quotedHaystack, quotedNeedle, strings.Repeat("^", len(quotedNeedle)))
	for _, rn := range rawPositions {
		if rn != '^' {
			positions += " "
		} else {
			positions += "^"
		}
	}
	return
}
