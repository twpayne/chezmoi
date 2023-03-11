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
	"github.com/google/go-github/v50/github"
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

const (
	upgradeMethodBrewUpgrade       = "brew-upgrade"
	upgradeMethodReplaceExecutable = "replace-executable"
	upgradeMethodSnapRefresh       = "snap-refresh"
	upgradeMethodUpgradePackage    = "upgrade-package"
	upgradeMethodSudoPrefix        = "sudo-"

	libcTypeGlibc = "glibc"
	libcTypeMusl  = "musl"

	packageTypeNone = ""
	packageTypeAPK  = "apk"
	packageTypeAUR  = "aur"
	packageTypeDEB  = "deb"
	packageTypeRPM  = "rpm"
)

var (
	packageTypeByID = map[string]string{
		"alpine":   packageTypeAPK,
		"amzn":     packageTypeRPM,
		"arch":     packageTypeAUR,
		"centos":   packageTypeRPM,
		"fedora":   packageTypeRPM,
		"opensuse": packageTypeRPM,
		"debian":   packageTypeDEB,
		"rhel":     packageTypeRPM,
		"sles":     packageTypeRPM,
		"ubuntu":   packageTypeDEB,
	}

	archReplacements = map[string]map[string]string{
		packageTypeDEB: {
			"386": "i386",
			"arm": "armel",
		},
		packageTypeRPM: {
			"amd64": "x86_64",
			"386":   "i686",
			"arm":   "armhfp",
			"arm64": "aarch64",
		},
	}

	checksumRx      = regexp.MustCompile(`\A([0-9a-f]{64})\s+(\S+)\z`)
	libcTypeGlibcRx = regexp.MustCompile(`(?i)glibc|gnu libc`)
	libcTypeMuslRx  = regexp.MustCompile(`(?i)musl`)
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
	flags.StringVar(&c.upgrade.executable, "executable", c.upgrade.method, "Set executable to replace")
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
		return errors.New("cannot upgrade dev version to latest released version unless --force is set")
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
		if err := c.upgradePackage(ctx, version, rr, useSudo); err != nil {
			return err
		}
	case upgradeMethodSudoPrefix + upgradeMethodUpgradePackage:
		useSudo := true
		if err := c.upgradePackage(ctx, version, rr, useSudo); err != nil {
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

func (c *Config) brewUpgrade() error {
	return c.run(chezmoi.EmptyAbsPath, "brew", []string{"upgrade", c.upgrade.repo})
}

func (c *Config) getChecksums(ctx context.Context, rr *github.RepositoryRelease) (map[string][]byte, error) {
	name := fmt.Sprintf("%s_%s_checksums.txt", c.upgrade.repo, strings.TrimPrefix(rr.GetTagName(), "v"))
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

// getLibc attempts to determine the system's libc.
func (c *Config) getLibc() (string, error) {
	// First, try parsing the output of ldd --version. On glibc systems it
	// writes to stdout and exits with code 0. On musl libc systems it writes to
	// stderr and exits with code 1.
	lddCmd := exec.Command("ldd", "--version")
	switch output, _ := chezmoilog.LogCmdCombinedOutput(lddCmd); {
	case libcTypeGlibcRx.Match(output):
		return libcTypeGlibc, nil
	case libcTypeMuslRx.Match(output):
		return libcTypeMusl, nil
	}

	// Second, try getconf GNU_LIBC_VERSION.
	getconfCmd := exec.Command("getconf", "GNU_LIBC_VERSION")
	if output, _ := chezmoilog.LogCmdCombinedOutput(getconfCmd); libcTypeGlibcRx.Match(output) {
		return libcTypeGlibc, nil
	}

	return "", errors.New("unable to determine libc")
}

func (c *Config) getPackageFilename(packageType string, version *semver.Version, os, arch string) (string, error) {
	if archReplacement, ok := archReplacements[packageType][arch]; ok {
		arch = archReplacement
	}
	switch packageType {
	case packageTypeAPK:
		return fmt.Sprintf("%s_%s_%s_%s.apk", c.upgrade.repo, version, os, arch), nil
	case packageTypeDEB:
		return fmt.Sprintf("%s_%s_%s_%s.deb", c.upgrade.repo, version, os, arch), nil
	case packageTypeRPM:
		return fmt.Sprintf("%s-%s-%s.rpm", c.upgrade.repo, version, arch), nil
	default:
		return "", fmt.Errorf("%s: unsupported package type", packageType)
	}
}

func (c *Config) replaceExecutable(
	ctx context.Context, executableFilenameAbsPath chezmoi.AbsPath, releaseVersion *semver.Version,
	rr *github.RepositoryRelease,
) (err error) {
	var archiveFormat chezmoi.ArchiveFormat
	var archiveName string
	switch {
	case runtime.GOOS == "windows":
		archiveFormat = chezmoi.ArchiveFormatZip
		archiveName = fmt.Sprintf("%s_%s_%s_%s.zip", c.upgrade.repo, releaseVersion, runtime.GOOS, runtime.GOARCH)
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		archiveFormat = chezmoi.ArchiveFormatTarGz
		var libc string
		if libc, err = c.getLibc(); err != nil {
			return
		}
		archiveName = fmt.Sprintf("%s_%s_%s-%s_%s.tar.gz", c.upgrade.repo, releaseVersion, runtime.GOOS, libc, runtime.GOARCH)
	case runtime.GOOS == "linux" && runtime.GOARCH == "386":
		archiveFormat = chezmoi.ArchiveFormatTarGz
		archiveName = fmt.Sprintf("%s_%s_%s_i386.tar.gz", c.upgrade.repo, releaseVersion, runtime.GOOS)
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
	walkArchiveFunc := func(name string, info fs.FileInfo, r io.Reader, linkname string) error {
		switch {
		case runtime.GOOS != "windows" && name == c.upgrade.repo:
			fallthrough
		case runtime.GOOS == "windows" && name == c.upgrade.repo+".exe":
			var err error
			executableData, err = io.ReadAll(r)
			if err != nil {
				return err
			}
			return chezmoi.Break
		default:
			return nil
		}
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
		if err = c.baseSystem.Rename(executableFilenameAbsPath, executableFilenameAbsPath.Append(".old")); err != nil {
			return
		}
	}
	err = c.baseSystem.WriteFile(executableFilenameAbsPath, executableData, 0o755)

	return
}

func (c *Config) snapRefresh() error {
	return c.run(chezmoi.EmptyAbsPath, "snap", []string{"refresh", c.upgrade.repo})
}

func (c *Config) upgradePackage(
	ctx context.Context, version *semver.Version, rr *github.RepositoryRelease, useSudo bool,
) error {
	switch runtime.GOOS {
	case "linux":
		// Determine the package type and architecture.
		packageType, err := getPackageType(c.baseSystem)
		if err != nil {
			return err
		}

		// chezmoi does not build and distribute AUR packages, so instead rely
		// on pacman and the community package.
		if packageType == packageTypeAUR {
			var args []string
			if useSudo {
				args = append(args, "sudo")
			}
			args = append(args, "pacman", "-S", c.upgrade.repo)
			return c.run(chezmoi.EmptyAbsPath, args[0], args[1:])
		}

		// Find the release asset.
		packageFilename, err := c.getPackageFilename(packageType, version, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return err
		}
		releaseAsset := getReleaseAssetByName(rr, packageFilename)
		if releaseAsset == nil {
			return fmt.Errorf("%s: cannot find release asset", packageFilename)
		}

		// Create a temporary directory for the package.
		tempDirAbsPath, err := c.tempDir("chezmoi")
		if err != nil {
			return err
		}

		data, err := c.downloadURL(ctx, releaseAsset.GetBrowserDownloadURL())
		if err != nil {
			return err
		}
		if err := c.verifyChecksum(ctx, rr, releaseAsset.GetName(), data); err != nil {
			return err
		}

		packageAbsPath := tempDirAbsPath.JoinString(releaseAsset.GetName())
		if err := c.baseSystem.WriteFile(packageAbsPath, data, 0o644); err != nil {
			return err
		}

		// Install the package from disk.
		var args []string
		if useSudo {
			args = append(args, "sudo")
		}
		switch packageType {
		case packageTypeAPK:
			args = append(args, "apk", "--allow-untrusted", packageAbsPath.String())
		case packageTypeDEB:
			args = append(args, "dpkg", "-i", packageAbsPath.String())
		case packageTypeRPM:
			args = append(args, "rpm", "-U", packageAbsPath.String())
		}
		return c.run(chezmoi.EmptyAbsPath, args[0], args[1:])
	default:
		return fmt.Errorf("%s: unsupported GOOS", runtime.GOOS)
	}
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
			"%s: checksum failed (want %s, got %s)", name, hex.EncodeToString(expectedChecksum), hex.EncodeToString(checksum[:]),
		)
	}
	return nil
}

// getUpgradeMethod attempts to determine the method by which chezmoi can be
// upgraded by looking at how it was installed.
func getUpgradeMethod(fileSystem vfs.Stater, executableAbsPath chezmoi.AbsPath) (string, error) {
	switch {
	case runtime.GOOS == "darwin" && strings.Contains(executableAbsPath.String(), "/homebrew/"):
		return upgradeMethodBrewUpgrade, nil
	case runtime.GOOS == "linux" && strings.Contains(executableAbsPath.String(), "/.linuxbrew/"):
		return upgradeMethodBrewUpgrade, nil
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

	switch runtime.GOOS {
	case "darwin":
		return upgradeMethodReplaceExecutable, nil
	case "freebsd":
		return upgradeMethodReplaceExecutable, nil
	case "linux":
		if ok, _ := vfs.Contains(fileSystem, executableAbsPath.String(), "/snap"); ok {
			return upgradeMethodSnapRefresh, nil
		}
		fileInfo, err := fileSystem.Stat(executableAbsPath.String())
		if err != nil {
			return "", err
		}
		uid := os.Getuid()
		switch fileInfoUID(fileInfo) {
		case 0:
			method := upgradeMethodUpgradePackage
			if uid != 0 {
				if _, err := chezmoi.LookPath("sudo"); err == nil {
					method = upgradeMethodSudoPrefix + method
				}
			}
			return method, nil
		case uid:
			return upgradeMethodReplaceExecutable, nil
		default:
			return "", fmt.Errorf("%s: cannot upgrade executable owned by non-current non-root user", executableAbsPath)
		}
	case "openbsd":
		return upgradeMethodReplaceExecutable, nil
	default:
		return "", nil
	}
}

// getPackageType returns the distributions package type based on is OS release.
func getPackageType(system chezmoi.System) (string, error) {
	osRelease, err := chezmoi.OSRelease(system.UnderlyingFS())
	if err != nil {
		return packageTypeNone, err
	}
	if id, ok := osRelease["ID"].(string); ok {
		if packageType, ok := packageTypeByID[id]; ok {
			return packageType, nil
		}
	}
	if idLikes, ok := osRelease["ID_LIKE"].(string); ok {
		for _, id := range strings.Split(idLikes, " ") {
			if packageType, ok := packageTypeByID[id]; ok {
				return packageType, nil
			}
		}
	}
	err = fmt.Errorf("could not determine package type (ID=%q, ID_LIKE=%q)", osRelease["ID"], osRelease["ID_LIKE"])
	return packageTypeNone, err
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
