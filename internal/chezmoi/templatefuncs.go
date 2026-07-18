package chezmoi

import "errors"

var (
	errReturnEmpty  = errors.New("return empty")
	errSkipTemplate = errors.New("skip template")
)

func AbortEmptyTemplateFunc() string {
	panic(errReturnEmpty)
}

func SkipTemplateIf(skip bool) {
	if skip {
		panic(errSkipTemplate)
	}
}
