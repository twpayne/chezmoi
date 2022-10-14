package chezmoi

// Delims are template delimiters.
type Delims struct {
	Left  string `mapstructure:"left"`
	Right string `mapstructure:"right"`
}
