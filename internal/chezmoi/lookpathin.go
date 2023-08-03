package chezmoi

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	foundExecutableCacheMutex sync.Mutex
	foundExecutableCache      = make(map[string]struct{})
)

// LookPathIn is like lookPath except that you can specify the paths rather than just using the current `$PATH`. This
// makes it useful for the resulting path of rc/profile files.
func LookPathIn(file, paths string) (string, error) {
	foundExecutableCacheMutex.Lock()
	defer foundExecutableCacheMutex.Unlock()

	// stolen from: /usr/lib/go-1.20/src/os/exec/lp_unix.go:52
	for _, dir := range filepath.SplitList(paths) {
		if dir == "" {
			continue
		}
		p := filepath.Join(dir, file)
		for _, path := range findExecutableExtensions(p) {
			if _, ok := foundExecutableCache[path]; ok {
				return path, nil
			}
			f, err := os.Stat(path)
			if err != nil {
				continue
			}
			m := f.Mode()
			// isExecutable doesn't care if it's a directory
			if m.IsDir() {
				continue
			}
			if isExecutable(f) {
				foundExecutableCache[path] = struct{}{}
				return path, nil
			}
		}
	}

	return "", nil
}
