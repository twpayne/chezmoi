package cmd

const (
	pathStyleAbsolute       = "absolute"
	pathStyleRelative       = "relative"
	pathStyleSourceAbsolute = "source-absolute"
	pathStyleSourceRelative = "source-relative"
	pathStyleAll            = "all"
)

var (
	sourceOrTargetPathStyleValues = []string{
		pathStyleAbsolute,
		pathStyleRelative,
		pathStyleSourceAbsolute,
		pathStyleSourceRelative,
		pathStyleAll,
	}
	targetPathStyleValues = []string{
		pathStyleAbsolute,
		pathStyleRelative,
	}
)
