package cmd

import (
	"os"
	"path/filepath"
)

func init() {
	config.addTemplateFunc("include", config.includeFunc)
	config.addTemplateFunc("stat", config.statFunc)
}

func (c *Config) includeFunc(filename string) string {
	contents, err := c.fs.ReadFile(filepath.Join(c.SourceDir, filename))
	if err != nil {
		panic(err)
	}
	return string(contents)
}

func (c *Config) statFunc(name string) interface{} {
	info, err := c.fs.Stat(name)
	switch {
	case err == nil:
		return map[string]interface{}{
			"name":    info.Name(),
			"size":    info.Size(),
			"mode":    int(info.Mode()),
			"perm":    int(info.Mode() & os.ModePerm),
			"modTime": info.ModTime().Unix(),
			"isDir":   info.IsDir(),
		}
	case os.IsNotExist(err):
		return nil
	default:
		panic(err)
	}
}
