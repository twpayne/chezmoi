package cmd

const (
	pathStyleAbsolute       = "absolute"
	pathStyleRelative       = "relative"
	pathStyleSourceAbsolute = "source-absolute"
	pathStyleSourceRelative = "source-relative"
)

var (
	sourceOrTargetPathStyleValues = []string{
		pathStyleAbsolute,
		pathStyleRelative,
		pathStyleSourceAbsolute,
		pathStyleSourceRelative,
	}
	targetPathStyleValues = []string{
		pathStyleAbsolute,
		pathStyleRelative,
	}
)
