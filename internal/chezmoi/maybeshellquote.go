package chezmoi

import (
	"regexp"
	"strings"
)

var needShellQuoteRegexp = regexp.MustCompile(`[^+\-./0-9=A-Z_a-z]`)

const (
	backslash   = '\\'
	singleQuote = '\''
)

// MaybeShellQuote returns s quoted as a shell argument, if necessary.
func MaybeShellQuote(s string) string {
	switch {
	case s == "":
		return "''"
	case needShellQuoteRegexp.MatchString(s):
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
				result = append(result, backslash, singleQuote)
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

// ShellQuoteArgs returs args shell quoted and joined into a single string.
func ShellQuoteArgs(args []string) string {
	shellQuotedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		shellQuotedArgs = append(shellQuotedArgs, MaybeShellQuote(arg))
	}
	return strings.Join(shellQuotedArgs, " ")
}
