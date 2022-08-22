package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/bradenhilton/mozillainstallhash"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"gopkg.in/ini.v1"
	"howett.net/plist"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

type ioregData struct {
	value map[string]any
}

var startOfLineRx = regexp.MustCompile(`(?m)^`)

func (c *Config) commentTemplateFunc(prefix, s string) string {
	return startOfLineRx.ReplaceAllString(s, prefix)
}

func (c *Config) fromIniTemplateFunc(s string) map[string]any {
	file, err := ini.Load([]byte(s))
	if err != nil {
		panic(err)
	}
	return iniFileToMap(file)
}

func (c *Config) fromTomlTemplateFunc(s string) any {
	var data any
	if err := chezmoi.FormatTOML.Unmarshal([]byte(s), &data); err != nil {
		panic(err)
	}
	return data
}

func (c *Config) fromYamlTemplateFunc(s string) any {
	var data any
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

func (c *Config) ioregTemplateFunc() map[string]any {
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

	var value map[string]any
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

func (c *Config) quoteListTemplateFunc(list []any) []string {
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

func (c *Config) replaceAllRegexTemplateFunc(expr, repl, s string) string {
	return regexp.MustCompile(expr).ReplaceAllString(s, repl)
}

func (c *Config) statTemplateFunc(name string) any {
	switch fileInfo, err := c.fileSystem.Stat(name); {
	case err == nil:
		return map[string]any{
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

func (c *Config) toIniTemplateFunc(data map[string]interface{}) string {
	var builder strings.Builder
	if err := writeIniMap(&builder, data, ""); err != nil {
		panic(err)
	}
	return builder.String()
}

func (c *Config) toTomlTemplateFunc(data any) string {
	toml, err := chezmoi.FormatTOML.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(toml)
}

func (c *Config) toYamlTemplateFunc(data any) string {
	yaml, err := chezmoi.FormatYAML.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(yaml)
}

func iniFileToMap(file *ini.File) map[string]any {
	m := make(map[string]any)
	for _, section := range file.Sections() {
		m[section.Name()] = iniSectionToMap(section)
	}
	return m
}

func iniSectionToMap(section *ini.Section) map[string]any {
	m := make(map[string]any)
	for _, s := range section.ChildSections() {
		m[s.Name()] = iniSectionToMap(s)
	}
	for _, k := range section.Keys() {
		m[k.Name()] = k.Value()
	}
	return m
}

func writeIniMap(w io.Writer, data map[string]any, sectionPrefix string) error {
	// Write keys in order and accumulate subsections.
	type subsection struct {
		key   string
		value map[string]any
	}
	var subsections []subsection
	for _, key := range sortedKeys(data) {
		switch value := data[key].(type) {
		case bool:
			fmt.Fprintf(w, "%s = %t\n", key, value)
		case float32, float64:
			fmt.Fprintf(w, "%s = %f\n", key, value)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			fmt.Fprintf(w, "%s = %d\n", key, value)
		case map[string]any:
			subsection := subsection{
				key:   key,
				value: value,
			}
			subsections = append(subsections, subsection)
		case string:
			fmt.Fprintf(w, "%s = %q\n", key, value)
		default:
			return fmt.Errorf("%s%s: %T: unsupported type", sectionPrefix, key, value)
		}
	}

	// Write subsections in order.
	for _, subsection := range subsections {
		if _, err := fmt.Fprintf(w, "\n[%s%s]\n", sectionPrefix, subsection.key); err != nil {
			return err
		}
		if err := writeIniMap(w, subsection.value, sectionPrefix+subsection.key+"."); err != nil {
			return err
		}
	}

	return nil
}

func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}
