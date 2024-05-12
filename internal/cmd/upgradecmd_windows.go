//go:build !noupgrade

package cmd

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v62/github"
	vfs "github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type InstallBehavior struct {
	PortablePackageUserRoot    string `json:"portablePackageUserRoot"`
	PortablePackageMachineRoot string `json:"portablePackageMachineRoot"`
}

func (ib *InstallBehavior) Values() []string {
	return []string{
		ib.PortablePackageUserRoot,
		ib.PortablePackageMachineRoot,
	}
}

type WinGetSettings struct {
	InstallBehavior InstallBehavior `json:"installBehavior"`
}

func (c *Config) brewUpgrade() error {
	return errUnsupportedUpgradeMethod
}

// isWinGetInstall determines if executableAbsPath contains a WinGet installation path.
func isWinGetInstall(fileSystem vfs.Stater, executableAbsPath string) (bool, error) {
	realExecutableAbsPath := executableAbsPath
	fi, err := os.Lstat(executableAbsPath)
	if err != nil {
		return false, err
	}
	if fi.Mode().Type() == fs.ModeSymlink {
		realExecutableAbsPath, err = os.Readlink(executableAbsPath)
		if err != nil {
			return false, err
		}
	}
	winGetSettings := WinGetSettings{
		InstallBehavior: InstallBehavior{
			PortablePackageUserRoot:    os.ExpandEnv(`${LOCALAPPDATA}\Microsoft\WinGet\Packages\`),
			PortablePackageMachineRoot: os.ExpandEnv(`${PROGRAMFILES}\WinGet\Packages\`),
		},
	}
	settingsPaths := []string{
		os.ExpandEnv(`${LOCALAPPDATA}\Packages\Microsoft.DesktopAppInstaller_8wekyb3d8bbwe\LocalState\settings.json`),
		os.ExpandEnv(`${LOCALAPPDATA}\Microsoft\WinGet\Settings\settings.json`),
	}
	for _, settingsPath := range settingsPaths {
		if _, err := os.Stat(settingsPath); err == nil {
			winGetSettingsContents, err := os.ReadFile(settingsPath)
			if err == nil {
				if err := chezmoi.FormatJSONC.Unmarshal(winGetSettingsContents, &winGetSettings); err != nil {
					return false, err
				}
			}
		}
	}
	for _, path := range winGetSettings.InstallBehavior.Values() {
		path = filepath.Clean(path)
		if path == "." {
			continue
		}
		if ok, _ := vfs.Contains(fileSystem, realExecutableAbsPath, path); ok {
			return true, nil
		}
	}
	return false, nil
}

func (c *Config) snapRefresh() error {
	return errUnsupportedUpgradeMethod
}

func (c *Config) upgradeUNIXPackage(
	ctx context.Context,
	version *semver.Version,
	rr *github.RepositoryRelease,
	useSudo bool,
) error {
	return errUnsupportedUpgradeMethod
}

func (c *Config) winGetUpgrade() error {
	return errors.New(
		"upgrade command is not currently supported for WinGet installations. chezmoi can still be upgraded via WinGet by running `winget upgrade --id twpayne.chezmoi --source winget`",
	)
}

// getLibc attempts to determine the system's libc.
func getLibc() (string, error) {
	return "", nil
}

// getUpgradeMethod attempts to determine the method by which chezmoi can be
// upgraded by looking at how it was installed.
func getUpgradeMethod(fileSystem vfs.Stater, executableAbsPath chezmoi.AbsPath) (string, error) {
	if ok, err := isWinGetInstall(fileSystem, executableAbsPath.String()); err != nil {
		return "", err
	} else if ok {
		return upgradeMethodWinGetUpgrade, nil
	}

	// If the executable is in the user's home directory, then always use
	// replace-executable.
	switch userHomeDir, err := chezmoi.UserHomeDir(); {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return "", err
	default:
		switch executableInUserHomeDir, err := vfs.Contains(fileSystem, executableAbsPath.String(), userHomeDir); {
		case errors.Is(err, fs.ErrNotExist):
		case err != nil:
			return "", err
		case executableInUserHomeDir:
			return upgradeMethodReplaceExecutable, nil
		}
	}

	// If the executable is in the system's temporary directory, then always use
	// replace-executable.
	if executableIsInTempDir, err := vfs.Contains(fileSystem, executableAbsPath.String(), os.TempDir()); err != nil {
		return "", err
	} else if executableIsInTempDir {
		return upgradeMethodReplaceExecutable, nil
	}

	return "", nil
}
