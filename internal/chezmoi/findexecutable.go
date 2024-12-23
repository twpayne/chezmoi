package chezmoi

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	foundExecutableCacheMutex sync.Mutex
	foundExecutableCache      = make(map[string]string)
)

// FindExecutable is like LookPath except that:
//
//   - You can specify the needle as `string`, `[]string`, or `[]interface{}`
//     (that converts to `[]string`).
//   - You specify the haystack instead of relying on `$PATH`/`%PATH%`.
//
// This makes it useful for the resulting path of shell configurations
// managed by chezmoi.
func FindExecutable(files []string, paths ...[]string) (string, error) {
	foundExecutableCacheMutex.Lock()
	defer foundExecutableCacheMutex.Unlock()
	if len(paths) == 0 || len(paths[0]) == 0 {
		paths = [][]string{filepath.SplitList(os.Getenv("PATH"))}
	}

	key := strings.Join(files, "\x00") + "\x01" + strings.Join(paths[0], "\x00")

	if path, ok := foundExecutableCache[key]; ok {
		return path, nil
	}

	var candidates []string

	for _, file := range files {
		candidates = append(candidates, findExecutableExtensions(file)...)
	}

	// based on /usr/lib/go-1.20/src/os/exec/lp_unix.go:52
	for _, candidatePath := range paths[0] {
		if candidatePath == "" {
			continue
		}

		for _, candidate := range candidates {
			path := filepath.Join(candidatePath, candidate)

			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			// isExecutable doesn't care if it's a directory
			if info.Mode().IsDir() {
				continue
			}

			if IsExecutable(info) {
				foundExecutableCache[key] = path
				return path, nil
	}

	return "", nil
}
