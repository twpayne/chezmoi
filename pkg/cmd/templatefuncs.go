package cmd

import (
	"encoding/hex"
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

// An emptyPathElementError is returned when a path element is empty.
type emptyPathElementError struct {
	index int
}

func (e emptyPathElementError) Error() string {
	return fmt.Sprintf("empty path element at index %d", e.index)
}

// An invalidPathElementTypeError is returned when an element in a path has an invalid type.
type invalidPathElementTypeError struct {
	element any
}

func (e invalidPathElementTypeError) Error() string {
	return fmt.Sprintf("%v: invalid path element type %T", e.element, e.element)
}

// An invalidPathTypeError is returned when a path has an invalid type.
type invalidPathTypeError struct {
	path any
}

func (e invalidPathTypeError) Error() string {
	return fmt.Sprintf("%v: invalid path type %T", e.path, e.path)
}

// errEmptyPath is returned when a path is empty.
var errEmptyPath = errors.New("empty path")

// needsQuoteRx matches any string that contains non-printable characters,
// double quotes, or a backslash.
var needsQuoteRx = regexp.MustCompile(`[^\x21\x23-\x5b\x5d-\x7e]`)

func (c *Config) commentTemplateFunc(prefix, s string) string {
	type stateType int
	const (
		startOfLine stateType = iota
		inLine
	)

	state := startOfLine
	var builder strings.Builder
	for _, r := range s {
		switch state {
		case startOfLine:
			if _, err := builder.WriteString(prefix); err != nil {
				panic(err)
			}
			if _, err := builder.WriteRune(r); err != nil {
				panic(err)
			}
			if r != '\n' {
				state = inLine
			}
		case inLine:
			if _, err := builder.WriteRune(r); err != nil {
				panic(err)
			}
			if r == '\n' {
				state = startOfLine
			}
		}
	}
	return builder.String()
}

func (c *Config) eqFoldTemplateFunc(first, second string, more ...string) bool {
	if strings.EqualFold(first, second) {
		return true
	}
	for _, s := range more {
		if strings.EqualFold(first, s) {
			return true
		}
	}
	return false
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

	matches, err := chezmoi.Glob(c.fileSystem, filepath.ToSlash(pattern))
	if err != nil {
		panic(err)
	}
	return matches
}

func (c *Config) hexDecodeTemplateFunc(s string) string {
	result, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func (c *Config) hexEncodeTemplateFunc(s string) string {
	return hex.EncodeToString([]byte(s))
}

func (c *Config) includeTemplateFunc(filename string) string {
	searchDirAbsPaths := []chezmoi.AbsPath{c.SourceDirAbsPath}
	contents, err := c.readFile(filename, searchDirAbsPaths)
	if err != nil {
		panic(err)
	}
	return string(contents)
}

func (c *Config) includeTemplateTemplateFunc(filename string, args ...any) string {
	var data any
	switch len(args) {
	case 0:
		// Do nothing.
	case 1:
		data = args[0]
	default:
		panic(fmt.Errorf("expected 0 or 1 arguments, got %d", len(args)))
	}

	searchDirAbsPaths := []chezmoi.AbsPath{
		c.SourceDirAbsPath.JoinString(chezmoi.TemplatesDirName),
		c.SourceDirAbsPath,
	}
	contents, err := c.readFile(filename, searchDirAbsPaths)
	if err != nil {
		panic(err)
	}

	templateOptions := chezmoi.TemplateOptions{
		Options: append([]string(nil), c.Template.Options...),
	}
	tmpl, err := chezmoi.ParseTemplate(filename, contents, c.templateFuncs, templateOptions)
	if err != nil {
		panic(err)
	}

	result, err := tmpl.Execute(data)
	if err != nil {
		panic(err)
	}
	return string(result)
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

func (c *Config) lstatTemplateFunc(name string) any {
	switch fileInfo, err := c.fileSystem.Lstat(name); {
	case err == nil:
		return fileInfoToMap(fileInfo)
	case errors.Is(err, fs.ErrNotExist):
		return nil
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

func (c *Config) readFile(filename string, searchDirAbsPaths []chezmoi.AbsPath) ([]byte, error) {
	if filepath.IsAbs(filename) {
		absPath, err := chezmoi.NewAbsPathFromExtPath(filename, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		return c.fileSystem.ReadFile(absPath.String())
	}

	var data []byte
	var err error
	for _, searchDir := range searchDirAbsPaths {
		data, err = c.fileSystem.ReadFile(searchDir.JoinString(filename).String())
		if !errors.Is(err, fs.ErrNotExist) {
			return data, err
		}
	}
	return data, err
}

func (c *Config) replaceAllRegexTemplateFunc(expr, repl, s string) string {
	return regexp.MustCompile(expr).ReplaceAllString(s, repl)
}

func (c *Config) setValueAtPathTemplateFunc(path, value, dict any) any {
	keys, lastKey, err := keysFromPath(path)
	if err != nil {
		panic(err)
	}

	result, ok := dict.(map[string]any)
	if !ok {
		result = make(map[string]any)
	}

	currentMap := result
	for _, key := range keys {
		if value, ok := currentMap[key]; ok {
			if nestedMap, ok := value.(map[string]any); ok {
				currentMap = nestedMap
			} else {
				nestedMap := make(map[string]any)
				currentMap[key] = nestedMap
				currentMap = nestedMap
			}
		} else {
			nestedMap := make(map[string]any)
			currentMap[key] = nestedMap
			currentMap = nestedMap
		}
	}
	currentMap[lastKey] = value

	return result
}

func (c *Config) statTemplateFunc(name string) any {
	switch fileInfo, err := c.fileSystem.Stat(name); {
	case err == nil:
		return fileInfoToMap(fileInfo)
	case errors.Is(err, fs.ErrNotExist):
		return nil
	default:
		panic(err)
	}
}

func (c *Config) toIniTemplateFunc(data map[string]any) string {
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

func fileInfoToMap(fileInfo fs.FileInfo) map[string]any {
	return map[string]any{
		"name":    fileInfo.Name(),
		"size":    fileInfo.Size(),
		"mode":    int(fileInfo.Mode()),
		"perm":    int(fileInfo.Mode().Perm()),
		"modTime": fileInfo.ModTime().Unix(),
		"isDir":   fileInfo.IsDir(),
		"type":    chezmoi.FileModeTypeNames[fileInfo.Mode()&fs.ModeType],
	}
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

func keysFromPath(path any) ([]string, string, error) {
	switch path := path.(type) {
	case string:
		if path == "" {
			return nil, "", errEmptyPath
		}
		keys := strings.Split(path, ".")
		for i, key := range keys {
			if key == "" {
				return nil, "", emptyPathElementError{
					index: i,
				}
			}
		}
		return keys[:len(keys)-1], keys[len(keys)-1], nil
	case []any:
		if len(path) == 0 {
			return nil, "", errEmptyPath
		}
		keys := make([]string, 0, len(path))
		for i, pathElement := range path {
			switch pathElementStr, ok := pathElement.(string); {
			case !ok:
				return nil, "", invalidPathElementTypeError{
					element: pathElement,
				}
			case pathElementStr == "":
				return nil, "", emptyPathElementError{
					index: i,
				}
			default:
				keys = append(keys, pathElementStr)
			}
		}
		return keys[:len(keys)-1], keys[len(keys)-1], nil
	case []string:
		if len(path) == 0 {
			return nil, "", errEmptyPath
		}
		for i, key := range path {
			if key == "" {
				return nil, "", emptyPathElementError{
					index: i,
				}
			}
		}
		return path[:len(path)-1], path[len(path)-1], nil
	case nil:
		return nil, "", errEmptyPath
	default:
		return nil, "", invalidPathTypeError{
			path: path,
		}
	}
}

func nestedMapAtPath(m map[string]any, path any) (map[string]any, string, error) {
	keys, lastKey, err := keysFromPath(path)
	if err != nil {
		return nil, "", err
	}
	for _, key := range keys {
		nestedMap, ok := m[key].(map[string]any)
		if !ok {
			return nil, "", nil
		}
		m = nestedMap
	}
	return m, lastKey, nil
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
			if needsQuote(value) {
				fmt.Fprintf(w, "%s = %q\n", key, value)
			} else {
				fmt.Fprintf(w, "%s = %s\n", key, value)
			}
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

func needsQuote(s string) bool {
	if s == "" {
		return true
	}
	if needsQuoteRx.MatchString(s) {
		return true
	}
	if _, err := strconv.ParseBool(s); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}
