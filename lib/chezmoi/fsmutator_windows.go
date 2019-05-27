// +build windows

package chezmoi

import (
    "path/filepath"
    "syscall"
    "os"

    "github.com/hectane/go-acl"
    "github.com/google/renameio"
    "golang.org/x/sys/windows"

    vfs "github.com/twpayne/go-vfs"
)

// TODO: this should probably just erase Everyone's permissions
func (a *FSMutator) MakePrivate(file string, umask os.FileMode) error {
    return a.Chmod(file, 0700&^umask)
}

func (a *FSMutator) IsPrivate(fi os.FileInfo, umask os.FileMode) bool {
    mode, err := acl.GetFileMode(fi.Name())
    if err != nil {
        return false
    }

    // on windows, a file is private if "Everyone" has no permissions on it.
    return (mode & 0007) == 0
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
    var RootPathName = filepath.VolumeName(fp) + "\\"

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
