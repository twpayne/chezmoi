package main

import (
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestLintData(t *testing.T) {
	for i, tc := range []struct {
		data           []byte
		expectedErrStr string
	}{
		{
			data:           nil,
			expectedErrStr: "",
		},
		{
			data:           []byte("package main\n"),
			expectedErrStr: "",
		},
		{
			data:           []byte("package main\r\n"),
			expectedErrStr: "CRLF line ending",
		},
		{
			data:           []byte("package main \n"),
			expectedErrStr: "trailing whitespace",
		},
		{
			data:           []byte("package main"),
			expectedErrStr: "no newline at end of file",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actualErr := lintData("main.go", tc.data)
			if tc.expectedErrStr == "" {
				assert.NoError(t, actualErr)
			} else {
				assert.Contains(t, actualErr.Error(), tc.expectedErrStr)
			}
		})
	}
}
