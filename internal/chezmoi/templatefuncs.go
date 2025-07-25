package chezmoi

import "errors"

var errReturnEmpty = errors.New("return empty")

func AbortEmptyTemplateFunc() string {
	panic(errReturnEmpty)
}
