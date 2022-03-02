package chezmoi

import (
	"os/exec"
	"sync"
)

type lookPathCacheValue struct {
	path string
	err  error
}

var (
	lookPathCacheMutex sync.Mutex
	lookPathCache      = make(map[string]lookPathCacheValue)
)

// LookPath is like os/exec.LookPath except that results are cached.
func LookPath(file string) (string, error) {
	lookPathCacheMutex.Lock()
	defer lookPathCacheMutex.Unlock()

	if result, ok := lookPathCache[file]; ok {
		return result.path, result.err
	}

	path, err := exec.LookPath(file)
	lookPathCache[file] = lookPathCacheValue{
		path: path,
		err:  err,
	}
	return path, err
}
