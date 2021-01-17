package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"howett.net/plist"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type ioregData struct {
	value map[string]interface{}
}

func (c *Config) includeTemplateFunc(filename string) string {
	contents, err := c.fs.ReadFile(string(c.sourceDirAbsPath.Join(chezmoi.RelPath(filename))))
	if err != nil {
		returnTemplateError(err)
		return ""
	}
	return string(contents)
}

func (c *Config) ioregTemplateFunc() map[string]interface{} {
	if runtime.GOOS != "darwin" {
		return nil
	}

	if c.ioregData.value != nil {
		return c.ioregData.value
	}

	cmd := exec.Command("ioreg", "-a", "-l")
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("ioreg: %w", err))
		return nil
	}

	var value map[string]interface{}
	if _, err := plist.Unmarshal(output, &value); err != nil {
		returnTemplateError(fmt.Errorf("ioreg: %w", err))
		return nil
	}
	c.ioregData.value = value
	return value
}

func (c *Config) joinPathTemplateFunc(elem ...string) string {
	return filepath.Join(elem...)
}

func (c *Config) lookPathTemplateFunc(file string) string {
	path, err := exec.LookPath(file)
	switch {
	case err == nil:
		return path
	case errors.Is(err, exec.ErrNotFound):
		return ""
	default:
		returnTemplateError(err)
		return ""
	}
}

func (c *Config) statTemplateFunc(name string) interface{} {
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
		returnTemplateError(err)
		return nil
	}
}

func returnTemplateError(err error) {
	panic(err)
}
