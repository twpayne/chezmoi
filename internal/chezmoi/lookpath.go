package chezmoi

import (
	"os/exec"
	"runtime"
	"sync"
)

var (
	lookPathCacheMutex sync.Mutex
	lookPathCache      = make(map[string]string)
)

// LookPath is like os/exec.LookPath except that the first positive result is
// cached on non-Windows systems.
func LookPath(file string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		// On Windows, defer to the Go standard library's os/exec.LookPath
		// function.
		return exec.LookPath(file)
	default:
		// On non-Windows systems, assume that the $PATH environment variable
		// doesn't change so we can cache results.
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
}
