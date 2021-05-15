// +build !noupgrade,!windows

package cmd

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v35/github"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs/v3"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

const (
	methodReplaceExecutable = "replace-executable"
	methodSnapRefresh       = "snap-refresh"
	methodUpgradePackage    = "upgrade-package"
	methodSudoPrefix        = "sudo-"

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
			"arm":   "armfp",
			"arm64": "aarch64",
		},
	}

	checksumRx      = regexp.MustCompile(`\A([0-9a-f]{64})\s+(\S+)\z`)
	libcTypeGlibcRx = regexp.MustCompile(`(?i)glibc|gnu libc`)
	libcTypeMuslRx  = regexp.MustCompile(`(?i)musl`)
)

func (c *Config) runUpgradeCmd(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if c.version == nil && !c.force {
		return errors.New("cannot upgrade dev version to latest released version unless --force is set")
	}

	client := newGitHubClient(ctx)

	// Get the latest release.
	rr, _, err := client.Repositories.GetLatestRelease(ctx, c.upgrade.owner, c.upgrade.repo)
	if err != nil {
		return err
	}
	releaseVersion, err := semver.NewVersion(strings.TrimPrefix(rr.GetName(), "v"))
	if err != nil {
		return err
	}

	// If the upgrade is not forced, stop if we're already the latest version.
	// Print a message and return no error so the command exits with success.
	if !c.force && !c.version.LessThan(*releaseVersion) {
		fmt.Fprintf(c.stdout, "chezmoi: already at the latest version (%s)\n", c.version)
		return nil
	}

	// Determine the upgrade method to use.
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	executableAbsPath := chezmoi.AbsPath(executable)
	method := c.upgrade.method
	if method == "" {
		method, err = getMethod(c.fileSystem, executableAbsPath)
		if err != nil {
			return err
		}
	}

	// Replace the executable with the updated version.
	switch method {
	case methodReplaceExecutable:
		if err := c.replaceExecutable(ctx, executableAbsPath, releaseVersion, rr); err != nil {
			return err
		}
	case methodSnapRefresh:
		if err := c.snapRefresh(); err != nil {
			return err
		}
	case methodUpgradePackage:
		useSudo := false
		if err := c.upgradePackage(ctx, rr, useSudo); err != nil {
			return err
		}
	case methodSudoPrefix + methodUpgradePackage:
		useSudo := true
		if err := c.upgradePackage(ctx, rr, useSudo); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: invalid --method value", method)
	}

	// Find the executable. If we replaced the executable directly, then use
	// that, otherwise look in $PATH.
	path := executable
	if method != methodReplaceExecutable {
		path, err = exec.LookPath(c.upgrade.repo)
		if err != nil {
			return err
		}
	}

	// Execute the new version.
	arg0 := path
	argv := []string{arg0, "--version"}
	log.Logger.Debug().
		Str("arg0", arg0).
		Strs("argv", argv).
		Msg("exec")
	err = syscall.EINTR
	for errors.Is(err, syscall.EINTR) {
		err = syscall.Exec(arg0, argv, os.Environ())
	}
	return err
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
		log.Logger.Error().
			Err(err).
			Str("url", url).
			Msg("http get")
		return nil, err
	}
	log.Logger.Debug().
		Str("url", url).
		Msg("http get")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: got a non-200 OK response: %d %s", url, resp.StatusCode, resp.Status)
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
	if output, _ := c.baseSystem.IdempotentCmdCombinedOutput(lddCmd); len(output) != 0 {
		switch {
		case libcTypeGlibcRx.Match(output):
			return libcTypeGlibc, nil
		case libcTypeMuslRx.Match(output):
			return libcTypeMusl, nil
		}
	}

	// Second, try getconf GNU_LIBC_VERSION.
	getconfCmd := exec.Command("getconf", "GNU_LIBC_VERSION")
	if output, err := c.baseSystem.IdempotentCmdOutput(getconfCmd); err != nil {
		if libcTypeGlibcRx.Match(output) {
			return libcTypeGlibc, nil
		}
	}

	return "", errors.New("unable to determine libc")
}

func (c *Config) replaceExecutable(ctx context.Context, executableFilenameAbsPath chezmoi.AbsPath, releaseVersion *semver.Version, rr *github.RepositoryRelease) error {
	goos := runtime.GOOS
	if goos == "linux" {
		libc, err := c.getLibc()
		if err != nil {
			return err
		}
		goos += "-" + libc
	}
	name := fmt.Sprintf("%s_%s_%s_%s.tar.gz", c.upgrade.repo, releaseVersion, goos, runtime.GOARCH)
	releaseAsset := getReleaseAssetByName(rr, name)
	if releaseAsset == nil {
		return fmt.Errorf("%s: cannot find release asset", name)
	}

	data, err := c.downloadURL(ctx, releaseAsset.GetBrowserDownloadURL())
	if err != nil {
		return err
	}
	if err := c.verifyChecksum(ctx, rr, releaseAsset.GetName(), data); err != nil {
		return err
	}

	// Extract the executable from the archive.
	gzipr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gzipr.Close()
	tr := tar.NewReader(gzipr)
	var executableData []byte
FOR:
	for {
		h, err := tr.Next()
		switch {
		case err == nil && h.Name == c.upgrade.repo:
			executableData, err = io.ReadAll(tr)
			if err != nil {
				return err
			}
			break FOR
		case errors.Is(err, io.EOF):
			return fmt.Errorf("%s: could not find header", c.upgrade.repo)
		}
	}

	return c.baseSystem.WriteFile(executableFilenameAbsPath, executableData, 0o755)
}

func (c *Config) snapRefresh() error {
	return c.run("", "snap", []string{"refresh", c.upgrade.repo})
}

func (c *Config) upgradePackage(ctx context.Context, rr *github.RepositoryRelease, useSudo bool) error {
	switch runtime.GOOS {
	case "darwin":
		return c.run("", "brew", []string{"upgrade", c.upgrade.repo})
	case "linux":
		// Determine the package type and architecture.
		packageType, err := getPackageType(c.fileSystem)
		if err != nil {
			return err
		}
		arch := runtime.GOARCH
		if archReplacement, ok := archReplacements[packageType]; ok {
			arch = archReplacement[arch]
		}

		// chezmoi does not build and distribute AUR packages, so instead rely
		// on pacman and the community package.
		if packageType == packageTypeAUR {
			var args []string
			if useSudo {
				args = append(args, "sudo")
			}
			args = append(args, "pacman", "-S", c.upgrade.repo)
			return c.run("", args[0], args[1:])
		}

		// Find the corresponding release asset.
		var releaseAsset *github.ReleaseAsset
		suffix := arch + "." + packageType
		for i, ra := range rr.Assets {
			if strings.HasSuffix(ra.GetName(), suffix) {
				releaseAsset = rr.Assets[i]
				break
			}
		}
		if releaseAsset == nil {
			return fmt.Errorf("cannot find release asset (arch=%q, packageType=%q)", arch, packageType)
		}

		// Create a temporary directory for the package.
		var tempDirAbsPath chezmoi.AbsPath
		if c.dryRun {
			tempDirAbsPath = chezmoi.AbsPath(os.TempDir())
		} else {
			tempDir, err := os.MkdirTemp("", "chezmoi")
			if err != nil {
				return err
			}
			tempDirAbsPath = chezmoi.AbsPath(tempDir)
			defer func() {
				_ = c.baseSystem.RemoveAll(tempDirAbsPath)
			}()
		}

		data, err := c.downloadURL(ctx, releaseAsset.GetBrowserDownloadURL())
		if err != nil {
			return err
		}
		if err := c.verifyChecksum(ctx, rr, releaseAsset.GetName(), data); err != nil {
			return err
		}

		packageFilename := tempDirAbsPath.Join(chezmoi.RelPath(releaseAsset.GetName()))
		if err := c.baseSystem.WriteFile(packageFilename, data, 0o644); err != nil {
			return err
		}

		// Install the package from disk.
		var args []string
		if useSudo {
			args = append(args, "sudo")
		}
		switch packageType {
		case packageTypeAPK:
			args = append(args, "apk", "--allow-untrusted", string(packageFilename))
		case packageTypeDEB:
			args = append(args, "dpkg", "-i", string(packageFilename))
		case packageTypeRPM:
			args = append(args, "rpm", "-U", string(packageFilename))
		}
		return c.run("", args[0], args[1:])
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
		return fmt.Errorf("%s: checksum failed (want %s, got %s)", name, hex.EncodeToString(expectedChecksum), hex.EncodeToString(checksum[:]))
	}
	return nil
}

// getMethod attempts to determine the method by which chezmoi can be upgraded
// by looking at how it was installed.
func getMethod(fileSystem vfs.Stater, executableAbsPath chezmoi.AbsPath) (string, error) {
	// If the executable is in the user's home directory, then always use
	// replace-executable.
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if executableInUserHomeDir, err := vfs.Contains(fileSystem, string(executableAbsPath), userHomeDir); err != nil {
		return "", err
	} else if executableInUserHomeDir {
		return methodReplaceExecutable, nil
	}

	// If the executable is in the system's temporary directory, then always use
	// replace-executable.
	if executableIsInTempDir, err := vfs.Contains(fileSystem, string(executableAbsPath), os.TempDir()); err != nil {
		return "", err
	} else if executableIsInTempDir {
		return methodReplaceExecutable, nil
	}

	switch runtime.GOOS {
	case "darwin":
		return methodUpgradePackage, nil
	case "freebsd":
		return methodReplaceExecutable, nil
	case "linux":
		if ok, _ := vfs.Contains(fileSystem, string(executableAbsPath), "/snap"); ok {
			return methodSnapRefresh, nil
		}

		info, err := fileSystem.Stat(string(executableAbsPath))
		if err != nil {
			return "", err
		}
		//nolint:forcetypeassert
		executableStat := info.Sys().(*syscall.Stat_t)
		uid := os.Getuid()
		switch int(executableStat.Uid) {
		case 0:
			method := methodUpgradePackage
			if uid != 0 {
				if _, err := exec.LookPath("sudo"); err == nil {
					method = methodSudoPrefix + method
				}
			}
			return method, nil
		case uid:
			return methodReplaceExecutable, nil
		default:
			return "", fmt.Errorf("%s: cannot upgrade executable owned by non-current non-root user", executableAbsPath)
		}
	case "openbsd":
		return methodReplaceExecutable, nil
	default:
		return "", fmt.Errorf("%s: unsupported GOOS", runtime.GOOS)
	}
}

func getPackageType(fileSystem vfs.FS) (string, error) {
	osRelease, err := chezmoi.OSRelease(fileSystem)
	if err != nil {
		return packageTypeNone, err
	}
	if id, ok := osRelease["ID"]; ok {
		if packageType, ok := packageTypeByID[id]; ok {
			return packageType, nil
		}
	}
	if idLikes, ok := osRelease["ID_LIKE"]; ok {
		for _, id := range strings.Split(idLikes, " ") {
			if packageType, ok := packageTypeByID[id]; ok {
				return packageType, nil
			}
		}
	}
	return packageTypeNone, fmt.Errorf("could not determine package type (ID=%q, ID_LIKE=%q)", osRelease["ID"], osRelease["ID_LIKE"])
}

func getReleaseAssetByName(rr *github.RepositoryRelease, name string) *github.ReleaseAsset {
	for i, ra := range rr.Assets {
		if ra.GetName() == name {
			return rr.Assets[i]
		}
	}
	return nil
}
