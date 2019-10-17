//+build windows

package cmd

import "strings"

func lines(s string) string {
	return strings.Replace(s, "\n", "\r\n", -1)
}
