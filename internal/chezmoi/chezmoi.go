// Package chezmoi contains chezmoi's core logic.
package chezmoi

import (
	"bufio"
	"bytes"
	"crypto/md5"  //nolint:gosec
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io/fs"
	"net"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs/v5"
	"golang.org/x/crypto/ripemd160" //nolint:gosec,staticcheck

	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

var (
	// DefaultTemplateOptions are the default template options.
	DefaultTemplateOptions = []string{"missingkey=error"}

	// Umask is the process's umask.
	Umask = fs.FileMode(0)
)

// Prefixes and suffixes.
const (
	ignorePrefix     = "."
	afterPrefix      = "after_"
	beforePrefix     = "before_"
	createPrefix     = "create_"
	dotPrefix        = "dot_"
	emptyPrefix      = "empty_"
	encryptedPrefix  = "encrypted_"
	exactPrefix      = "exact_"
	executablePrefix = "executable_"
	externalPrefix   = "external_"
	literalPrefix    = "literal_"
	modifyPrefix     = "modify_"
	oncePrefix       = "once_"
	onChangePrefix   = "onchange_"
	privatePrefix    = "private_"
	readOnlyPrefix   = "readonly_"
	removePrefix     = "remove_"
	runPrefix        = "run_"
	symlinkPrefix    = "symlink_"
	literalSuffix    = ".literal"
	TemplateSuffix   = ".tmpl"
)

// Special file names.
const (
	Prefix = ".chezmoi"

	RootName         = Prefix + "root"
	TemplatesDirName = Prefix + "templates"
	VersionName      = Prefix + "version"
	dataName         = Prefix + "data"
	externalName     = Prefix + "external"
	externalsDirName = Prefix + "externals"
	ignoreName       = Prefix + "ignore"
	removeName       = Prefix + "remove"
	scriptsDirName   = Prefix + "scripts"
)

var (
	dirPrefixRx  = regexp.MustCompile(`\A(dot|exact|literal|readonly|private)_`)
	filePrefixRx = regexp.MustCompile(
		`\A(after|before|create|dot|empty|encrypted|executable|literal|modify|once|private|readonly|remove|run|symlink)_`,
	)
	fileSuffixRx = regexp.MustCompile(`\.(literal|tmpl)\z`)
	whitespaceRx = regexp.MustCompile(`\s+`)
)

// knownPrefixedFiles is a set of known filenames with the .chezmoi prefix.
var knownPrefixedFiles = chezmoiset.New(
	Prefix+".json"+TemplateSuffix,
	Prefix+".toml"+TemplateSuffix,
	Prefix+".yaml"+TemplateSuffix,
	RootName,
	VersionName,
	dataName+".json",
	dataName+".toml",
	dataName+".yaml",
	externalName+".json"+TemplateSuffix,
	externalName+".json",
	externalName+".toml"+TemplateSuffix,
	externalName+".toml",
	externalName+".yaml"+TemplateSuffix,
	externalName+".yaml",
	ignoreName+TemplateSuffix,
	ignoreName,
	removeName+TemplateSuffix,
	removeName,
)

// knownPrefixedDirs is a set of known dirnames with the .chezmoi prefix.
var knownPrefixedDirs = chezmoiset.New(
	TemplatesDirName,
	dataName,
	externalsDirName,
	scriptsDirName,
)

// knownTargetFiles is a set of known target files that should not be managed
// directly.
var knownTargetFiles = chezmoiset.New(
	"chezmoi.json",
	"chezmoi.toml",
	"chezmoi.yaml",
	"chezmoistate.boltdb",
)

var FileModeTypeNames = map[fs.FileMode]string{
	0:                 "file",
	fs.ModeDir:        "dir",
	fs.ModeSymlink:    "symlink",
	fs.ModeNamedPipe:  "named pipe",
	fs.ModeSocket:     "socket",
	fs.ModeDevice:     "device",
	fs.ModeCharDevice: "char device",
}

// FQDNHostname returns the FQDN hostname.
func FQDNHostname(fileSystem vfs.FS) (string, error) {
	// First, try os.Hostname. If it returns something that looks like a FQDN
	// hostname, or we're on Windows, return it.
	osHostname, err := os.Hostname()
	if runtime.GOOS == "windows" || (err == nil && strings.Contains(osHostname, ".")) {
		return osHostname, err
	}

	// Otherwise, if we're on OpenBSD, try /etc/myname.
	if runtime.GOOS == "openbsd" {
		if fqdnHostname, err := etcMynameFQDNHostname(fileSystem); err == nil && fqdnHostname != "" {
			return fqdnHostname, nil
		}
	}

	// Otherwise, try /etc/hosts.
	if fqdnHostname, err := etcHostsFQDNHostname(fileSystem); err == nil && fqdnHostname != "" {
		return fqdnHostname, nil
	}

	// Otherwise, try /etc/hostname.
	if fqdnHostname, err := etcHostnameFQDNHostname(fileSystem); err == nil && fqdnHostname != "" {
		return fqdnHostname, nil
	}

	// Finally, fall back to whatever os.Hostname returned.
	return osHostname, err
}

// FlagCompletionFunc returns a flag completion function.
func FlagCompletionFunc(allCompletions []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		for _, completion := range allCompletions {
			if strings.HasPrefix(completion, toComplete) {
				completions = append(completions, completion)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// ParseBool is like strconv.ParseBool but also accepts on, ON, y, Y, yes, YES,
// n, N, no, NO, off, and OFF.
func ParseBool(str string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(str)) {
	case "n", "no", "off":
		return false, nil
	case "on", "y", "yes":
		return true, nil
	default:
		return strconv.ParseBool(str)
	}
}

// SuspiciousSourceDirEntry returns true if base is a suspicious dir entry.
func SuspiciousSourceDirEntry(base string, fileInfo fs.FileInfo, encryptedSuffixes []string) bool {
	switch fileInfo.Mode().Type() {
	case 0:
		if strings.HasPrefix(base, Prefix) && !knownPrefixedFiles.Contains(base) {
			return true
		}
		for _, encryptedSuffix := range encryptedSuffixes {
			fileAttr := parseFileAttr(fileInfo.Name(), encryptedSuffix)
			if knownTargetFiles.Contains(fileAttr.TargetName) {
				return true
			}
		}
		return false
	case fs.ModeDir:
		return strings.HasPrefix(base, Prefix) && !knownPrefixedDirs.Contains(base)
	case fs.ModeSymlink:
		return strings.HasPrefix(base, Prefix)
	default:
		return true
	}
}

// UniqueAbbreviations returns a map of unique abbreviations of values to
// values. Values always map to themselves.
func UniqueAbbreviations(values []string) map[string]string {
	abbreviations := make(map[string][]string)
	for _, value := range values {
		for i := 1; i <= len(value); i++ {
			abbreviation := value[:i]
			abbreviations[abbreviation] = append(abbreviations[abbreviation], value)
		}
	}
	uniqueAbbreviations := make(map[string]string)
	for abbreviation, values := range abbreviations {
		if len(values) == 1 {
			uniqueAbbreviations[abbreviation] = values[0]
		}
	}
	for _, value := range values {
		uniqueAbbreviations[value] = value
	}
	return uniqueAbbreviations
}

// eagerNoErr returns a function that returns an eagerly-evaluated value and no
// error.
func eagerNoErr[T any](value T) func() (T, error) {
	return func() (T, error) {
		return value, nil
	}
}

// etcHostnameFQDNHostname returns the FQDN hostname from parsing /etc/hostname.
func etcHostnameFQDNHostname(fileSystem vfs.FS) (string, error) {
	contents, err := fileSystem.ReadFile("/etc/hostname")
	if err != nil {
		return "", err
	}
	s := bufio.NewScanner(bytes.NewReader(contents))
	for s.Scan() {
		text := s.Text()
		text, _, _ = strings.Cut(text, "#")
		if hostname := strings.TrimSpace(text); hostname != "" {
			return hostname, nil
		}
	}
	return "", s.Err()
}

// etcMynameFQDNHostname returns the FQDN hostname from parsing /etc/myname.
// See OpenBSD's myname(5) for details on this file.
func etcMynameFQDNHostname(fileSystem vfs.FS) (string, error) {
	contents, err := fileSystem.ReadFile("/etc/myname")
	if err != nil {
		return "", err
	}
	s := bufio.NewScanner(bytes.NewReader(contents))
	for s.Scan() {
		text := s.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}
		if hostname := strings.TrimSpace(text); hostname != "" {
			return hostname, nil
		}
	}
	return "", s.Err()
}

// etcHostsFQDNHostname returns the FQDN hostname from parsing /etc/hosts.
func etcHostsFQDNHostname(fileSystem vfs.FS) (string, error) {
	contents, err := fileSystem.ReadFile("/etc/hosts")
	if err != nil {
		return "", err
	}
	s := bufio.NewScanner(bytes.NewReader(contents))
	for s.Scan() {
		text := s.Text()
		text = strings.TrimSpace(text)
		text, _, _ = strings.Cut(text, "#")
		fields := whitespaceRx.Split(text, -1)
		if len(fields) < 2 {
			continue
		}
		if !net.ParseIP(fields[0]).IsLoopback() {
			continue
		}
		hostname, domainname, found := strings.Cut(fields[1], ".")
		if !found {
			continue
		}
		if hostname == "localhost" {
			continue
		}
		if domainname == "localdomain" {
			continue
		}
		// Docker Desktop breaks /etc/hosts. Filter out all docker.internal
		// domain names. See https://github.com/twpayne/chezmoi/issues/3095.
		if domainname == "docker.internal" {
			continue
		}
		if runtime.GOOS == "darwin" && domainname == "local" {
			continue
		}
		return fields[1], nil
	}
	return "", s.Err()
}

// isEmpty returns true if data is empty after trimming whitespace from both
// ends.
func isEmpty(data []byte) bool {
	return len(bytes.TrimSpace(data)) == 0
}

// isPrivate returns if fileInfo is private.
func isPrivate(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o77 == 0
}

// isReadOnly returns if fileInfo is read-only.
func isReadOnly(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o222 == 0
}

// md5Sum returns the MD5 sum of data.
func md5Sum(data []byte) []byte {
	md5SumArr := md5.Sum(data) //nolint:gosec
	return md5SumArr[:]
}

// lazySHA256 returns a function that returns a SHA256 computed lazily.
func lazySHA256(contentsFunc func() ([]byte, error)) func() ([32]byte, error) {
	return sync.OnceValues(func() ([32]byte, error) {
		contents, err := contentsFunc()
		if err != nil {
			return [32]byte{}, err
		}
		return sha256.Sum256(contents), nil
	})
}

// modeTypeName returns a string representation of mode.
func modeTypeName(mode fs.FileMode) string {
	if name, ok := FileModeTypeNames[mode.Type()]; ok {
		return name
	}
	return fmt.Sprintf("0o%o: unknown type", mode.Type())
}

// ripemd160Sum returns the RIPEMD-160 sum of data.
func ripemd160Sum(data []byte) []byte {
	return ripemd160.New().Sum(data) //nolint:gosec
}

// sha1Sum returns the SHA1 sum of data.
func sha1Sum(data []byte) []byte {
	sha1SumArr := sha1.Sum(data) //nolint:gosec
	return sha1SumArr[:]
}

// sha384Sum returns the SHA384 sum of data.
func sha384Sum(data []byte) []byte {
	sha384SumArr := sha512.Sum384(data)
	return sha384SumArr[:]
}

// sha512Sum returns the SHA512 sum of data.
func sha512Sum(data []byte) []byte {
	sha512SumArr := sha512.Sum512(data)
	return sha512SumArr[:]
}

// ensureSuffix adds suffix to s if s is not suffixed by suffix.
func ensureSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}
