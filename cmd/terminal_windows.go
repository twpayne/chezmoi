package cmd

import (
	"io"
)

func getWriterSupportsColor(w io.Writer) bool {
	return false
}

func getWriterWidth(w io.Writer) int {
	return 80
}
