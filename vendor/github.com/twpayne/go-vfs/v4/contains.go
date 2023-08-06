package vfs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// A Stater implements Stat. It is assumed that the fs.FileInfos returned by
// Stat are compatible with os.SameFile.
type Stater interface {
	Stat(string) (fs.FileInfo, error)
}

// Contains returns true if p is reachable by traversing through prefix. prefix
// must exist, but p may not. It is an expensive but accurate alternative to the
// deprecated filepath.HasPrefix.
func Contains(fileSystem Stater, p, prefix string) (bool, error) {
	prefixFI, err := fileSystem.Stat(prefix)
	if err != nil {
		return false, err
	}
	for {
		fi, err := fileSystem.Stat(p)
		switch {
		case err == nil:
			if os.SameFile(fi, prefixFI) {
				return true, nil
			}
			goto TryParent
		case errors.Is(err, fs.ErrNotExist):
			goto TryParent
		case errors.Is(err, fs.ErrPermission):
			goto TryParent
		default:
			// Remove any fs.PathError or os.SyscallError wrapping, if present.
		Unwrap:
			for {
				var pathError *fs.PathError
				var syscallError *os.SyscallError
				switch {
				case errors.As(err, &pathError):
					err = pathError.Err
				case errors.As(err, &syscallError):
					err = syscallError.Err
				default:
					break Unwrap
				}
			}
			// Ignore some syscall.Errnos.
			var syscallErrno syscall.Errno
			if errors.As(err, &syscallErrno) {
				if _, ignore := ignoreErrnoInContains[syscallErrno]; ignore {
					goto TryParent
				}
			}
			return false, err
		}
	TryParent:
		parentDir := filepath.Dir(p)
		if parentDir == p {
			// Return when we stop making progress.
			return false, nil
		}
		p = parentDir
	}
}
