package chezmoi

import (
	"os/exec"
	"strings"
	"sync"
)

var (
	lookPathCacheMutex sync.Mutex
	lookPathCache      = make(map[string]string)
)

// LookPath is like os/exec.LookPath except that the first positive result is
// cached.
func LookPath(file string) (string, error) {
	return LookOnePath([]string{file})
}

// LookOnePath is like os/exec.LookPath where any one of the provided files
// will be returned and the first positive result is cached.
func LookOnePath(files []string) (string, error) {
	lookPathCacheMutex.Lock()
	defer lookPathCacheMutex.Unlock()

	key := strings.Join(files, "\x00")

	if path, ok := lookPathCache[key]; ok {
		return path, nil
	}

	for _, file := range files {
		path, err := exec.LookPath(file)
		if err == nil {
			lookPathCache[key] = path

			return path, nil
		}
	}

	return "", exec.ErrNotFound
}
