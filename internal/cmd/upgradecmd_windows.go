//go:build !noupgrade

package cmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v54/github"
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

const (
	upgradeMethodReplaceExecutable = "replace-executable"
	upgradeMethodWinGetUpgrade     = "winget-upgrade"
)

var checksumRx = regexp.MustCompile(`\A([0-9a-f]{64})\s+(\S+)\z`)

type upgradeCmdConfig struct {
	executable string
	method     string
	owner      string
	repo       string
}

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

func (c *Config) newUpgradeCmd() *cobra.Command {
	upgradeCmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade chezmoi to the latest released version",
		Long:    mustLongHelp("upgrade"),
		Example: example("upgrade"),
		Args:    cobra.NoArgs,
		RunE:    c.runUpgradeCmd,
		Annotations: newAnnotations(
			runsCommands,
		),
	}

	flags := upgradeCmd.Flags()
	flags.StringVar(
		&c.upgrade.executable,
		"executable",
		c.upgrade.method,
		"Set executable to replace",
	)
	flags.StringVar(&c.upgrade.method, "method", c.upgrade.method, "Set upgrade method")
	flags.StringVar(&c.upgrade.owner, "owner", c.upgrade.owner, "Set owner")
	flags.StringVar(&c.upgrade.repo, "repo", c.upgrade.repo, "Set repo")

	return upgradeCmd
}

func (c *Config) runUpgradeCmd(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var zeroVersion semver.Version
	if c.version == zeroVersion && !c.force {
		return errors.New(
			"cannot upgrade dev version to latest released version unless --force is set",
		)
	}

	httpClient, err := c.getHTTPClient()
	if err != nil {
		return err
	}
	client := chezmoi.NewGitHubClient(ctx, httpClient)

	// Get the latest release.
	rr, _, err := client.Repositories.GetLatestRelease(ctx, c.upgrade.owner, c.upgrade.repo)
	if err != nil {
		return err
	}
	version, err := semver.NewVersion(strings.TrimPrefix(rr.GetName(), "v"))
	if err != nil {
		return err
	}

	// If the upgrade is not forced, stop if we're already the latest version.
	// Print a message and return no error so the command exits with success.
	if !c.force && !c.version.LessThan(*version) {
		fmt.Fprintf(c.stdout, "chezmoi: already at the latest version (%s)\n", c.version)
		return nil
	}

	// Determine the upgrade method to use.
	if c.upgrade.executable == "" {
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		c.upgrade.executable = executable
	}

	executableAbsPath := chezmoi.NewAbsPath(c.upgrade.executable)
	method := c.upgrade.method
	if method == "" {
		switch method, err = getUpgradeMethod(c.fileSystem, executableAbsPath); {
		case err != nil:
			return err
		case method == "":
			return fmt.Errorf(
				"%s/%s: cannot determine upgrade method for %s",
				runtime.GOOS,
				runtime.GOARCH,
				executableAbsPath,
			)
		}
	}
	c.logger.Info().
		Str("executable", c.upgrade.executable).
		Str("method", method).
		Msg("upgradeMethod")

	// Replace the executable with the updated version.
	switch method {
	case upgradeMethodReplaceExecutable:
		if err := c.replaceExecutable(ctx, executableAbsPath, version, rr); err != nil {
			return err
		}
	case upgradeMethodWinGetUpgrade:
		if err := c.winGetUpgrade(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: invalid method", method)
	}

	// Find the executable. If we replaced the executable directly, then use
	// that, otherwise look in $PATH.
	path := c.upgrade.executable
	if method != upgradeMethodReplaceExecutable {
		path, err = chezmoi.LookPath(c.upgrade.repo)
		if err != nil {
			return err
		}
	}

	// Execute the new version.
	chezmoiVersionCmd := exec.Command(path, "--version")
	chezmoiVersionCmd.Stdin = os.Stdin
	chezmoiVersionCmd.Stdout = os.Stdout
	chezmoiVersionCmd.Stderr = os.Stderr
	return chezmoilog.LogCmdRun(chezmoiVersionCmd)
}

func (c *Config) getChecksums(
	ctx context.Context,
	rr *github.RepositoryRelease,
) (map[string][]byte, error) {
	name := fmt.Sprintf(
		"%s_%s_checksums.txt",
		c.upgrade.repo,
		strings.TrimPrefix(rr.GetTagName(), "v"),
	)
	releaseAsset := getReleaseAssetByName(rr, name)
	if releaseAsset == nil {
		return nil, fmt.Errorf("%s: cannot find release asset", name)
	}

	data, err := c.downloadURL(ctx, releaseAsset.GetBrowserDownloadURL())
	if err != nil {
		return nil, err
	}

	checksums := make(map[string][]byte)
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		m := checksumRx.FindStringSubmatch(s.Text())
		if m == nil {
			return nil, fmt.Errorf("%q: cannot parse checksum", s.Text())
		}
		checksums[m[2]], _ = hex.DecodeString(m[1])
	}
	return checksums, s.Err()
}

func (c *Config) downloadURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	httpClient, err := c.getHTTPClient()
	if err != nil {
		return nil, err
	}
	resp, err := chezmoilog.LogHTTPRequest(c.logger, httpClient, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Config) replaceExecutable(
	ctx context.Context, executableFilenameAbsPath chezmoi.AbsPath, releaseVersion *semver.Version,
	rr *github.RepositoryRelease,
) (err error) {
	var archiveFormat chezmoi.ArchiveFormat
	var archiveName string
	archiveFormat = chezmoi.ArchiveFormatZip
	archiveName = fmt.Sprintf(
		"%s_%s_%s_%s.zip",
		c.upgrade.repo,
		releaseVersion,
		runtime.GOOS,
		runtime.GOARCH,
	)
	releaseAsset := getReleaseAssetByName(rr, archiveName)
	if releaseAsset == nil {
		err = fmt.Errorf("%s: cannot find release asset", archiveName)
		return
	}

	var archiveData []byte
	if archiveData, err = c.downloadURL(ctx, releaseAsset.GetBrowserDownloadURL()); err != nil {
		return
	}
	if err = c.verifyChecksum(ctx, rr, releaseAsset.GetName(), archiveData); err != nil {
		return
	}

	// Extract the executable from the archive.
	var executableData []byte
	walkArchiveFunc := func(name string, info fs.FileInfo, r io.Reader, linkname string) error {
		if name == c.upgrade.repo+".exe" {
			var err error
			executableData, err = io.ReadAll(r)
			if err != nil {
				return err
			}
			return fs.SkipAll
		}
		return nil
	}
	if err = chezmoi.WalkArchive(archiveData, archiveFormat, walkArchiveFunc); err != nil {
		return
	}
	if executableData == nil {
		err = fmt.Errorf("%s: cannot find executable in archive", archiveName)
		return
	}

	// Replace the executable.
	if err = c.baseSystem.Rename(executableFilenameAbsPath, executableFilenameAbsPath.Append(".old")); err != nil {
		return
	}
	err = c.baseSystem.WriteFile(executableFilenameAbsPath, executableData, 0o755)

	return
}

func (c *Config) verifyChecksum(
	ctx context.Context,
	rr *github.RepositoryRelease,
	name string,
	data []byte,
) error {
	checksums, err := c.getChecksums(ctx, rr)
	if err != nil {
		return err
	}
	expectedChecksum, ok := checksums[name]
	if !ok {
		return fmt.Errorf("%s: checksum not found", name)
	}
	checksum := sha256.Sum256(data)
	if !bytes.Equal(checksum[:], expectedChecksum) {
		return fmt.Errorf(
			"%s: checksum failed (want %s, got %s)",
			name,
			hex.EncodeToString(expectedChecksum),
			hex.EncodeToString(checksum[:]),
		)
	}
	return nil
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

func (c *Config) winGetUpgrade() error {
	return fmt.Errorf("upgrade command is not currently supported for WinGet installations. chezmoi can still be upgraded via WinGet by running `winget upgrade --id %s.%s --source winget`", c.upgrade.owner, c.upgrade.repo)
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
	switch userHomeDir, err := os.UserHomeDir(); {
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

// getReleaseAssetByName returns the release asset from rr with the given name.
func getReleaseAssetByName(rr *github.RepositoryRelease, name string) *github.ReleaseAsset {
	for i, ra := range rr.Assets {
		if ra.GetName() == name {
			return rr.Assets[i]
		}
	}
	return nil
}
