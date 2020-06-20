package cmd

import "path/filepath"

func init() {
	config.addTemplateFunc("include", config.includeFunc)
}

func (c *Config) includeFunc(filename string) string {
	contents, err := c.fs.ReadFile(filepath.Join(c.SourceDir, filename))
	if err != nil {
		panic(err)
	}
	return string(contents)
}
