package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/bradenhilton/mozillainstallhash"
	"howett.net/plist"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

type ioregData struct {
	value map[string]interface{}
}

var startOfLineRx = regexp.MustCompile(`(?m)^`)

func (c *Config) commentTemplateFunc(prefix, s string) string {
	return startOfLineRx.ReplaceAllString(s, prefix)
}

func (c *Config) fromTomlTemplateFunc(s string) interface{} {
	var data interface{}
	if err := chezmoi.FormatTOML.Unmarshal([]byte(s), &data); err != nil {
		panic(err)
	}
	return data
}

func (c *Config) fromYamlTemplateFunc(s string) interface{} {
	var data interface{}
	if err := chezmoi.FormatYAML.Unmarshal([]byte(s), &data); err != nil {
		panic(err)
	}
	return data
}

func (c *Config) globTemplateFunc(pattern string) []string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defer func() {
		value := recover()
		err := os.Chdir(wd)
		if value != nil {
			panic(value)
		}
		if err != nil {
			panic(err)
		}
	}()

	if err := os.Chdir(c.DestDirAbsPath.String()); err != nil {
		panic(err)
	}

	matches, err := doublestar.Glob(c.fileSystem, filepath.ToSlash(pattern))
	if err != nil {
		panic(err)
	}
	return matches
}

func (c *Config) includeTemplateFunc(filename string) string {
	var absPath chezmoi.AbsPath
	if filepath.IsAbs(filename) {
		var err error
		absPath, err = chezmoi.NewAbsPathFromExtPath(filename, c.homeDirAbsPath)
		if err != nil {
			panic(err)
		}
	} else {
		absPath = c.SourceDirAbsPath.JoinString(filename)
	}
	contents, err := c.fileSystem.ReadFile(absPath.String())
	if err != nil {
		panic(err)
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

	command := "ioreg"
	args := []string{"-a", "-l"}
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}

	var value map[string]interface{}
	if _, err := plist.Unmarshal(output, &value); err != nil {
		panic(newParseCmdOutputError(command, args, output, err))
	}
	c.ioregData.value = value
	return value
}

func (c *Config) joinPathTemplateFunc(elem ...string) string {
	return filepath.Join(elem...)
}

func (c *Config) lookPathTemplateFunc(file string) string {
	switch path, err := chezmoi.LookPath(file); {
	case err == nil:
		return path
	case errors.Is(err, exec.ErrNotFound):
		return ""
	case errors.Is(err, fs.ErrNotExist):
		return ""
	default:
		panic(err)
	}
}

func (c *Config) mozillaInstallHashTemplateFunc(path string) string {
	mozillaInstallHash, err := mozillainstallhash.MozillaInstallHash(path)
	if err != nil {
		panic(err)
	}
	return mozillaInstallHash
}

func (c *Config) outputTemplateFunc(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}
	// FIXME we should be able to return output directly, but
	// github.com/Masterminds/sprig's trim function only accepts strings
	return string(output)
}

func (c *Config) quoteListTemplateFunc(list []interface{}) []string {
	result := make([]string, 0, len(list))
	for _, elem := range list {
		var elemStr string
		switch elem := elem.(type) {
		case []byte:
			elemStr = string(elem)
		case string:
			elemStr = elem
		case error:
			elemStr = elem.Error()
		case fmt.Stringer:
			elemStr = elem.String()
		default:
			elemStr = fmt.Sprintf("%v", elem)
		}
		result = append(result, strconv.Quote(elemStr))
	}
	return result
}

func (c *Config) statTemplateFunc(name string) interface{} {
	switch fileInfo, err := c.fileSystem.Stat(name); {
	case err == nil:
		return map[string]interface{}{
			"name":    fileInfo.Name(),
			"size":    fileInfo.Size(),
			"mode":    int(fileInfo.Mode()),
			"perm":    int(fileInfo.Mode().Perm()),
			"modTime": fileInfo.ModTime().Unix(),
			"isDir":   fileInfo.IsDir(),
		}
	case errors.Is(err, fs.ErrNotExist):
		return nil
	default:
		panic(err)
	}
}

func (c *Config) toTomlTemplateFunc(data interface{}) string {
	toml, err := chezmoi.FormatTOML.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(toml)
}

func (c *Config) toYamlTemplateFunc(data interface{}) string {
	yaml, err := chezmoi.FormatYAML.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(yaml)
}
