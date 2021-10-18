package cmd

import (
	"regexp"
	"strings"
)

// nonShellLiteralRx is a regular expression that matches anything that is not a
// shell literal.
var nonShellLiteralRx = regexp.MustCompile(`[^+\-./0-9=A-Z_a-z]`)

// shellQuote returns s quoted as a shell argument, if necessary.
func shellQuote(s string) string {
	const (
		backslash   = '\\'
		singleQuote = '\''
	)

	switch {
	case s == "":
		return "''"
	case nonShellLiteralRx.MatchString(s):
		result := make([]byte, 0, 2+len(s))
		inSingleQuotes := false
		for _, b := range []byte(s) {
			switch b {
			case backslash:
				if !inSingleQuotes {
					result = append(result, singleQuote)
					inSingleQuotes = true
				}
				result = append(result, backslash, backslash)
			case singleQuote:
				if inSingleQuotes {
					result = append(result, singleQuote)
					inSingleQuotes = false
				}
				result = append(result, '\\', singleQuote)
			default:
				if !inSingleQuotes {
					result = append(result, singleQuote)
					inSingleQuotes = true
				}
				result = append(result, b)
			}
		}
		if inSingleQuotes {
			result = append(result, singleQuote)
		}
		return string(result)
	default:
		return s
	}
}

// shellQuoteCommand returns a string containing command and args shell quoted.
func shellQuoteCommand(command string, args []string) string {
	if len(args) == 0 {
		return shellQuote(command)
	}
	elems := make([]string, 0, 1+len(args))
	elems = append(elems, shellQuote(command))
	for _, arg := range args {
		elems = append(elems, shellQuote(arg))
	}
	return strings.Join(elems, " ")
}
