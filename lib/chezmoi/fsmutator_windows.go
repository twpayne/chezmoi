// +build windows

package chezmoi

import (
	"os"
	"path/filepath"
	"syscall"

	acl "github.com/hectane/go-acl"

	"github.com/google/renameio"
	"golang.org/x/sys/windows"

	vfs "github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"
)

func (a *FSMutator) getCorrectedPath(file string) string {
	var pathfs *vfs.PathFS

	// if it just is a pathfs, cool
	maybe_pathfs, ok := a.FS.(*vfs.PathFS)
	if !ok {
		maybe_testfs, ok := a.FS.(*vfst.TestFS)

		if ok {
			pathfs = &maybe_testfs.PathFS
		}
	} else {
		pathfs = maybe_pathfs
	}

	if pathfs != nil {
		joined, err := pathfs.Join("", file)
		if err == nil {
			file = joined
		}
	}

	return file
}

// Use the same default value as Linux (as of kernel 4.19)
const MAXSYMLINKS = 40

func (a *FSMutator) IsPrivate(file string, umask os.FileMode) bool {
	file = a.getCorrectedPath(file)

	// if file is a symlink, get the path it links to.  this emulates
	// unix-style behavior, where symlinks can't have their own independent
	// permissions.

	resolved := file
	for i := 0; i < MAXSYMLINKS; i++ {
		fi, err := os.Lstat(resolved)
		if err != nil {
			return false
		}

		if fi.Mode()&os.ModeSymlink == 0 {
			// not a link, all done
			break
		}

		next, err := os.Readlink(resolved)
		if err != nil {
			return false
		}

		if next != "" && !filepath.IsAbs(next) {
			resolved = filepath.Join(filepath.Dir(resolved), next)
		}
	}

	mode, err := acl.GetEffectiveAccessMode(resolved)
	if err != nil {
		return false
	}

	return (uint32(mode) & 0007) == 0
}

// WriteFile implements Mutator.WriteFile.
func (a *FSMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if a.FS == vfs.OSFS {
		dir := filepath.Dir(name)
		dev, ok := a.devCache[dir]
		if !ok {
			volumeID, err := getVolumeSerialNumber(name)
			if err != nil {
				return err
			}

			dev = volumeID
			a.devCache[dir] = dev
		}
		tempDir, ok := a.tempDirCache[dev]
		if !ok {
			tempDir = renameio.TempDir(dir)
			a.tempDirCache[dev] = tempDir
		}
		t, err := renameio.TempFile(tempDir, name)
		if err != nil {
			return err
		}
		defer func() {
			_ = t.Cleanup()
		}()
		if err := a.Chmod(t.Name(), perm); err != nil {
			return err
		}
		if _, err := t.Write(data); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}
	return a.FS.WriteFile(name, data, perm)
}

func getVolumeSerialNumber(Path string) (uint, error) {
	fp, err := filepath.Abs(Path)
	if err != nil {
		return 0, err
	}

	// Input rootpath
	RootPathName := filepath.VolumeName(fp) + "\\"

	// Output volume info
	var serialNumber uint32

	err = windows.GetVolumeInformation(
		syscall.StringToUTF16Ptr(RootPathName),
		nil, 0,
		&serialNumber,
		nil,    // maximum component length
		nil,    // filesystem flags
		nil, 0, // filesystem name buffer
	)

	if err != windows.Errno(0) {
		return 0, err
	}

	return uint(serialNumber), nil
}
