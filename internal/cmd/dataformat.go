package cmd

const (
	formatUnknown = ""
	formatJSON    = "json"
	formatTOML    = "toml"
	formatYAML    = "yaml"
)

var (
	readDataFormatValues = []string{
		formatUnknown,
		formatJSON,
		formatTOML,
		formatYAML,
	}
	writeDataFormatValues = []string{
		formatUnknown,
		formatJSON,
		formatYAML,
	}
)
