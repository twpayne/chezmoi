// Package chezmoiassert implements testing assertions not implemented by
// github.com/alecthomas/assert/v2.
package chezmoiassert

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func PanicsWithError(tb testing.TB, expected error, fn func(), msgAndArgs ...any) {
	tb.Helper()
	defer func() {
		if value, ok := recover().(error); ok {
			assert.Equal(tb, expected, value, msgAndArgs...)
		} else {
			msg := formatMsgAndArgs("Expected function to panic with error", msgAndArgs...)
			tb.Fatal(msg)
		}
	}()
	fn()
}

func PanicsWithErrorString(tb testing.TB, errString string, fn func(), msgAndArgs ...any) {
	tb.Helper()
	defer func() {
		if value, ok := recover().(error); ok {
			assert.EqualError(tb, value, errString, msgAndArgs...)
		} else {
			msg := formatMsgAndArgs("Expected function to panic with error string", msgAndArgs...)
			tb.Fatal(msg)
		}
	}()
	fn()
}

func formatMsgAndArgs(dflt string, msgAndArgs ...any) string {
	if len(msgAndArgs) == 0 {
		return dflt
	}
	return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) //nolint:forcetypeassert
}
