package cmd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/bradenhilton/mozillainstallhash"
	"github.com/itchyny/gojq"
	"gopkg.in/ini.v1"
	"howett.net/plist"

	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/chezmoilog"
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
			_ = mustValue(builder.WriteString(prefix))
			_ = mustValue(builder.WriteRune(r))
			if r != '\n' {
				state = inLine
			}
		case inLine:
			_ = mustValue(builder.WriteRune(r))
			if r == '\n' {
				state = startOfLine
			}
		}
	}
	return builder.String()
}

func (c *Config) deleteValueAtPathTemplateFunc(path string, dict map[string]any) any {
	keys, lastKey := mustValues(keysFromPath(path))
	currentMap := dict
	for _, key := range keys {
		value, ok := currentMap[key]
		if !ok {
			return dict
		}
		nestedMap, ok := value.(map[string]any)
		if !ok {
			return dict
		}
		currentMap = nestedMap
	}
	delete(currentMap, lastKey)

	return dict
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

func (c *Config) findExecutableTemplateFunc(file string, pathList any) string {
	files := []string{file}
	paths, err := anyToStringSlice(pathList)
	if err != nil {
		panic(fmt.Errorf("path list: %w", err))
	}

	path, err := chezmoi.FindExecutable(files, paths)
	if err != nil {
		panic(err)
	}
	return path
}

func (c *Config) findOneExecutableTemplateFunc(fileList, pathList any) string {
	files, err := anyToStringSlice(fileList)
	if err != nil {
		panic(fmt.Errorf("file list: %w", err))
	}

	paths, err := anyToStringSlice(pathList)
	if err != nil {
		panic(fmt.Errorf("path list: %w", err))
	}

	path, err := chezmoi.FindExecutable(files, paths)
	if err != nil {
		panic(err)
	}
	return path
}

func (c *Config) fromIniTemplateFunc(s string) map[string]any {
	return iniFileToMap(mustValue(ini.Load([]byte(s))))
}

// fromJsonTemplateFunc parses s as JSON and returns the result. In contrast to
// encoding/json, numbers are represented as int64s or float64s if possible.
//
//nolint:revive,staticcheck
func (c *Config) fromJsonTemplateFunc(s string) any {
	var value any
	must(chezmoi.FormatJSON.Unmarshal([]byte(s), &value))
	return value
}

// fromJsoncTemplateFunc parses s as JSONC and returns the result. In contrast
// to encoding/json, numbers are represented as int64s or float64s if possible.
func (c *Config) fromJsoncTemplateFunc(s string) any {
	var value any
	must(chezmoi.FormatJSONC.Unmarshal([]byte(s), &value))
	return value
}

func (c *Config) fromTomlTemplateFunc(s string) any {
	var value map[string]any
	must(chezmoi.FormatTOML.Unmarshal([]byte(s), &value))
	return value
}

func (c *Config) fromYamlTemplateFunc(s string) any {
	var value any
	must(chezmoi.FormatYAML.Unmarshal([]byte(s), &value))
	return value
}

func (c *Config) getRedirectedURLTemplateFunc(requestURL string) string {
	client, err := c.getHTTPClient()
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodHead, requestURL, http.NoBody)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return resp.Request.URL.String()
}

func (c *Config) globTemplateFunc(pattern string) []string {
	defer func() {
		value := recover()
		err := os.Chdir(c.commandDirAbsPath.String())
		if value != nil {
			panic(value)
		}
		if err != nil {
			panic(err)
		}
	}()

	must(os.Chdir(c.DestDirAbsPath.String()))

	return mustValue(chezmoi.Glob(c.fileSystem, filepath.ToSlash(pattern)))
}

func (c *Config) hexDecodeTemplateFunc(s string) string {
	return string(mustValue(hex.DecodeString(s)))
}

func (c *Config) hexEncodeTemplateFunc(s string) string {
	return hex.EncodeToString([]byte(s))
}

func (c *Config) includeTemplateFunc(filename string) string {
	searchDirAbsPaths := []chezmoi.AbsPath{c.sourceDirAbsPath}
	return string(mustValue(c.readFile(filename, searchDirAbsPaths)))
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
		c.sourceDirAbsPath.JoinString(chezmoi.TemplatesDirName),
		c.sourceDirAbsPath,
	}
	contents := mustValue(c.readFile(filename, searchDirAbsPaths))

	tmpl := mustValue(chezmoi.ParseTemplate(filename, contents, chezmoi.TemplateOptions{
		Funcs:   c.templateFuncs,
		Options: slices.Clone(c.Template.Options),
	}))

	return string(mustValue(tmpl.Execute(data)))
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
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}

	var value map[string]any
	_ = mustValue(plist.Unmarshal(output, &value))
	c.ioregData.value = value
	return value
}

func (c *Config) joinPathTemplateFunc(elem ...string) string {
	return filepath.Join(elem...)
}

func (c *Config) jqTemplateFunc(source string, input any) any {
	query := mustValue(gojq.Parse(source))
	code := mustValue(gojq.Compile(query))
	iter := code.Run(input)
	var result []any
	for {
		value, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := value.(error); ok {
			panic(err)
		}
		result = append(result, value)
	}
	return result
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

func (c *Config) isExecutableTemplateFunc(file string) bool {
	switch fileInfo, err := c.fileSystem.Stat(file); {
	case err == nil:
		return chezmoi.IsExecutable(fileInfo)
	case errors.Is(err, fs.ErrNotExist):
		return false
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
	return mustValue(mozillainstallhash.MozillaInstallHash(path))
}

func (c *Config) outputTemplateFunc(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}
	return string(output)
}

func (c *Config) outputListTemplateFunc(name string, args []any) string {
	slice, err := anyToStringSlice(args)
	if err != nil {
		panic(fmt.Errorf("args list: %w", err))
	}

	return c.outputTemplateFunc(name, slice...)
}

func (c *Config) pruneEmptyDictsTemplateFunc(dict map[string]any) map[string]any {
	pruneEmptyMaps(dict)
	return dict
}

func (c *Config) quoteTemplateFunc(list ...any) string {
	ss := make([]string, 0, len(list))
	for _, elem := range list {
		if elem != nil {
			ss = append(ss, strconv.Quote(anyToString(elem)))
		}
	}
	return strings.Join(ss, " ")
}

func (c *Config) squoteTemplateFunc(list ...any) string {
	ss := make([]string, 0, len(list))
	for _, elem := range list {
		if elem != nil {
			ss = append(ss, "'"+anyToString(elem)+"'")
		}
	}
	return strings.Join(ss, " ")
}

func (c *Config) quoteListTemplateFunc(list []any) []string {
	result := make([]string, len(list))
	for i, elem := range list {
		result[i] = strconv.Quote(anyToString(elem))
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
	keys, lastKey := mustValues(keysFromPath(path))

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

func (c *Config) splitListTemplateFunc(sep, s string) []any {
	strSlice := strings.Split(s, sep)
	result := make([]any, len(strSlice))
	for i, v := range strSlice {
		result[i] = v
	}
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
	must(writeIniMap(&builder, data, ""))
	return builder.String()
}

func (c *Config) toPrettyJsonTemplateFunc(args ...any) string { //nolint:revive,staticcheck
	var (
		indent = "  "
		value  any
	)
	switch len(args) {
	case 1:
		value = args[0]
	case 2:
		var ok bool
		indent, ok = args[0].(string)
		if !ok {
			panic(fmt.Errorf("arg 1: expected a string, got a %T", args[0]))
		}
		value = args[1]
	default:
		panic(fmt.Errorf("expected 1 or 2 arguments, got %d", len(args)))
	}

	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)
	must(encoder.Encode(value))
	return builder.String()
}

func (c *Config) toStringTemplateFunc(value any) string {
	return anyToString(value)
}

func (c *Config) toStringsTemplateFunc(values ...any) []string {
	result, err := anyToStringSlice(values)
	if err != nil {
		panic(err)
	}
	return result
}

func (c *Config) toTomlTemplateFunc(data any) string {
	return string(mustValue(chezmoi.FormatTOML.Marshal(data)))
}

func (c *Config) toYamlTemplateFunc(data any) string {
	return string(mustValue(chezmoi.FormatYAML.Marshal(data)))
}

func (c *Config) warnfTemplateFunc(format string, args ...any) string {
	c.errorf("warning: "+format+"\n", args...)
	return ""
}

func anyToString(value any) string {
	switch value := value.(type) {
	case bool:
		return strconv.FormatBool(value)
	case *bool:
		if value == nil {
			return "false"
		}
		return strconv.FormatBool(*value)
	case []byte:
		return string(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case *float64:
		if value == nil {
			return "0"
		}
		return strconv.FormatFloat(*value, 'f', -1, 64)
	case int:
		return strconv.Itoa(value)
	case *int:
		if value == nil {
			return "0"
		}
		return strconv.Itoa(*value)
	case string:
		return value
	case *string:
		if value == nil {
			return ""
		}
		return *value
	case error:
		return value.Error()
	case fmt.Stringer:
		return value.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func anyToStringSlice(value any) ([]string, error) {
	switch value := value.(type) {
	case []any:
		result := make([]string, 0, len(value))
		for _, elem := range value {
			elemStrs, err := anyToStringSlice(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, elemStrs...)
		}
		return result, nil
	case []bool:
		result := make([]string, len(value))
		for i, value := range value {
			result[i] = strconv.FormatBool(value)
		}
		return result, nil
	case []float64:
		result := make([]string, len(value))
		for i, value := range value {
			result[i] = strconv.FormatFloat(value, 'f', -1, 64)
		}
		return result, nil
	case []int:
		result := make([]string, len(value))
		for i, value := range value {
			result[i] = strconv.Itoa(value)
		}
		return result, nil
	case []string:
		return value, nil
	default:
		return []string{anyToString(value)}, nil
	}
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
		if section.Name() == ini.DefaultSection {
			for _, k := range section.Keys() {
				m[k.Name()] = k.Value()
			}
		} else {
			m[section.Name()] = iniSectionToMap(section)
		}
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
		keys := make([]string, len(path))
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
				keys[i] = pathElementStr
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
	for _, key := range slices.Sorted(maps.Keys(data)) {
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
			fmt.Fprintf(w, "%s = %s\n", key, maybeQuote(value))
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

func maybeQuote(s string) string {
	if needsQuote(s) {
		return strconv.Quote(s)
	}
	return s
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

// pruneEmptyMaps prunes all empty maps from m and returns if m is now empty
// itself.
func pruneEmptyMaps(m map[string]any) bool {
	for key, value := range m {
		nestedMap, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if pruneEmptyMaps(nestedMap) {
			delete(m, key)
		}
	}
	return len(m) == 0
}
