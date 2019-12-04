package chezmoi

import (
	"regexp"
)

var needShellQuoteRegexp = regexp.MustCompile(`[^+\-./0-9A-Z_a-z]`)

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
