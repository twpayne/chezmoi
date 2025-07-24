package chezmoi

import "errors"

var errReturnEmpty = errors.New("return empty")

func ReturnEmptyTemplateFunc() string {
	panic(errReturnEmpty)
}
