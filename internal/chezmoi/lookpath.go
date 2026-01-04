package chezmoi

import (
	"os/exec"
	"sync"
)

var (
	lookPathCacheMutex sync.Mutex
	lookPathCache      = make(map[string]string)
)

// LookPath is like [os/exec.LookPath] except that the first positive result is
// cached.
func LookPath(file string) (string, error) {
	lookPathCacheMutex.Lock()
	defer lookPathCacheMutex.Unlock()

	if path, ok := lookPathCache[file]; ok {
		return path, nil
	}

	path, err := exec.LookPath(file)
	if err == nil {
		lookPathCache[file] = path
	}

	return path, err
}
