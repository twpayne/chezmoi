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
	"regexp"
	"runtime"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v57/github"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

const (
	upgradeMethodBrewUpgrade       = "brew-upgrade"
	upgradeMethodReplaceExecutable = "replace-executable"
	upgradeMethodSnapRefresh       = "snap-refresh"
	upgradeMethodUpgradePackage    = "upgrade-package"
	upgradeMethodSudoPrefix        = "sudo-"
	upgradeMethodWinGetUpgrade     = "winget-upgrade"
)

var (
	checksumRx                  = regexp.MustCompile(`\A([0-9a-f]{64})\s+(\S+)\z`)
	errUnsupportedUpgradeMethod = errors.New("unsupported upgrade method")
)

type upgradeCmdConfig struct {
	executable string
	method     string
	owner      string
	repo       string
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
			return fmt.Errorf("%s/%s: cannot determine upgrade method for %s", runtime.GOOS, runtime.GOARCH, executableAbsPath)
		}
	}
	c.logger.Info().
		Str("executable", c.upgrade.executable).
		Str("method", method).
		Msg("upgradeMethod")

	// Replace the executable with the updated version.
	switch method {
	case upgradeMethodBrewUpgrade:
		if err := c.brewUpgrade(); err != nil {
			return err
		}
	case upgradeMethodReplaceExecutable:
		if err := c.replaceExecutable(ctx, executableAbsPath, version, rr); err != nil {
			return err
		}
	case upgradeMethodSnapRefresh:
		if err := c.snapRefresh(); err != nil {
			return err
		}
	case upgradeMethodUpgradePackage:
		useSudo := false
		if err := c.upgradeUNIXPackage(ctx, version, rr, useSudo); err != nil {
			return err
		}
	case upgradeMethodSudoPrefix + upgradeMethodUpgradePackage:
		useSudo := true
		if err := c.upgradeUNIXPackage(ctx, version, rr, useSudo); err != nil {
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

func (c *Config) getChecksums(ctx context.Context, rr *github.RepositoryRelease) (map[string][]byte, error) {
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
	ctx context.Context,
	executableFilenameAbsPath chezmoi.AbsPath,
	releaseVersion *semver.Version,
	rr *github.RepositoryRelease,
) (err error) {
	var archiveFormat chezmoi.ArchiveFormat
	var archiveName string
	switch {
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		archiveFormat = chezmoi.ArchiveFormatTarGz
		var libc string
		if libc, err = getLibc(); err != nil {
			return
		}
		archiveName = fmt.Sprintf("%s_%s_%s-%s_%s.tar.gz", c.upgrade.repo, releaseVersion, runtime.GOOS, libc, runtime.GOARCH)
	case runtime.GOOS == "linux" && runtime.GOARCH == "386":
		archiveFormat = chezmoi.ArchiveFormatTarGz
		archiveName = fmt.Sprintf("%s_%s_%s_i386.tar.gz", c.upgrade.repo, releaseVersion, runtime.GOOS)
	case runtime.GOOS == "windows":
		archiveFormat = chezmoi.ArchiveFormatZip
		archiveName = fmt.Sprintf("%s_%s_%s_%s.zip", c.upgrade.repo, releaseVersion, runtime.GOOS, runtime.GOARCH)
	default:
		archiveFormat = chezmoi.ArchiveFormatTarGz
		archiveName = fmt.Sprintf("%s_%s_%s_%s.tar.gz", c.upgrade.repo, releaseVersion, runtime.GOOS, runtime.GOARCH)
	}
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
	executableName := c.upgrade.repo
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	walkArchiveFunc := func(name string, info fs.FileInfo, r io.Reader, linkname string) error {
		if name == executableName {
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
	if runtime.GOOS == "windows" {
		if err := c.baseSystem.Rename(executableFilenameAbsPath, executableFilenameAbsPath.Append(".old")); err != nil {
			return err
		}
	}
	err = c.baseSystem.WriteFile(executableFilenameAbsPath, executableData, 0o755)

	return
}

func (c *Config) verifyChecksum(ctx context.Context, rr *github.RepositoryRelease, name string, data []byte) error {
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

// getReleaseAssetByName returns the release asset from rr with the given name.
func getReleaseAssetByName(rr *github.RepositoryRelease, name string) *github.ReleaseAsset {
	for i, ra := range rr.Assets {
		if ra.GetName() == name {
			return rr.Assets[i]
		}
	}
	return nil
}
