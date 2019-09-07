// +build !noupgrade
// +build !windows

package cmd

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v26/github"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
	"golang.org/x/oauth2"
)

const (
	methodReplaceExecutable = "replace-executable"
	methodSnapRefresh       = "snap-refresh"
	methodUpgradePackage    = "upgrade-package"
	methodSudoPrefix        = "sudo-"

	packageTypeNone = ""
	packageTypeAUR  = "aur"
	packageTypeDEB  = "deb"
	packageTypeRPM  = "rpm"
)

var (
	packageTypeByID = map[string]string{
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

	checksumRegexp = regexp.MustCompile(`\A([0-9a-f]{64})\s+(\S+)\z`)
)

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Args:    cobra.NoArgs,
	Short:   "Upgrade chezmoi",
	Long:    mustGetLongHelp("upgrade"),
	Example: getExample("upgrade"),
	RunE:    makeRunE(config.runUpgradeCmd),
}

type upgradeCmdConfig struct {
	force  bool
	method string
	owner  string
	repo   string
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	persistentFlags := upgradeCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.upgrade.force, "force", "f", false, "force upgrade")
	persistentFlags.StringVarP(&config.upgrade.method, "method", "m", "", "set method")
	persistentFlags.StringVarP(&config.upgrade.owner, "owner", "o", "twpayne", "set owner")
	persistentFlags.StringVarP(&config.upgrade.repo, "repo", "r", "chezmoi", "set repo")
}

func (c *Config) runUpgradeCmd(fs vfs.FS, args []string) error {
	ctx := context.Background()

	// Use a GitHub API token, if set.
	var httpClient *http.Client
	if accessToken, ok := os.LookupEnv(strings.ToUpper(c.upgrade.repo) + "_GITHUB_API_TOKEN"); ok {
		httpClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: accessToken,
		}))
	}

	client := github.NewClient(httpClient)

	// Get the latest release.
	rr, _, err := client.Repositories.GetLatestRelease(ctx, c.upgrade.owner, c.upgrade.repo)
	if err != nil {
		return err
	}
	releaseVersion, err := semver.NewVersion(strings.TrimPrefix(rr.GetName(), "v"))
	if err != nil {
		return err
	}

	// If the upgrade is not forced and we're not a dev version, stop if we're
	// already the latest version.
	if !c.upgrade.force && VersionStr != devVersionStr {
		if !Version.LessThan(*releaseVersion) {
			fmt.Fprintf(c.Stdout(), "chezmoi: already at the latest version (%s)\n", Version)
			return nil
		}
	}

	// Determine the upgrade method to use.
	executableFilename, err := os.Executable()
	if err != nil {
		return err
	}
	method := c.upgrade.method
	if method == "" {
		method, err = getMethod(fs, executableFilename)
		if err != nil {
			return err
		}
	}

	// Replace the executable with the updated version.
	mutator := c.getDefaultMutator(fs)
	switch method {
	case methodReplaceExecutable:
		if err := c.replaceExecutable(mutator, executableFilename, releaseVersion, rr); err != nil {
			return err
		}
	case methodSnapRefresh:
		if err := c.snapRefresh(fs); err != nil {
			return err
		}
	case methodUpgradePackage:
		if err := c.upgradePackage(fs, mutator, rr, false); err != nil {
			return err
		}
	case methodSudoPrefix + methodUpgradePackage:
		if err := c.upgradePackage(fs, mutator, rr, true); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid --method value: %s", method)
	}

	// Find the executable. If we replaced the executable directly, then use
	// that, otherwise look in $PATH.
	path := executableFilename
	if method != methodReplaceExecutable {
		path, err = exec.LookPath(c.upgrade.repo)
		if err != nil {
			return err
		}
	}

	// Execute the new version.
	if c.Verbose {
		fmt.Printf("exec %s --version\n", path)
	}
	return syscall.Exec(path, []string{path, "--version"}, os.Environ())
}

func (c *Config) getChecksums(rr *github.RepositoryRelease) (map[string][]byte, error) {
	name := "checksums.txt"
	releaseAsset := getReleaseAssetByName(rr, name)
	if releaseAsset == nil {
		return nil, fmt.Errorf("%s: cannot find release asset", name)
	}

	data, err := c.downloadURL(releaseAsset.GetBrowserDownloadURL())
	if err != nil {
		return nil, err
	}

	checksums := make(map[string][]byte)
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		m := checksumRegexp.FindStringSubmatch(s.Text())
		if m == nil {
			return nil, fmt.Errorf("%q: cannot parse checksum", s.Text())
		}
		checksums[m[2]], _ = hex.DecodeString(m[1])
	}
	return checksums, s.Err()
}

func (c *Config) downloadURL(url string) ([]byte, error) {
	if c.Verbose {
		fmt.Fprintf(c.Stdout(), "curl -s -L %s\n", url)
	}
	//nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: got a non-200 OK response: %d %s", url, resp.StatusCode, resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Config) replaceExecutable(mutator chezmoi.Mutator, executableFilename string, releaseVersion *semver.Version, rr *github.RepositoryRelease) error {
	name := fmt.Sprintf("%s_%s_%s_%s.tar.gz", c.upgrade.repo, releaseVersion, runtime.GOOS, runtime.GOARCH)
	releaseAsset := getReleaseAssetByName(rr, name)
	if releaseAsset == nil {
		return fmt.Errorf("%s: cannot find release asset", name)
	}

	data, err := c.downloadURL(releaseAsset.GetBrowserDownloadURL())
	if err != nil {
		return err
	}
	if err := c.verifyChecksum(rr, releaseAsset.GetName(), data); err != nil {
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
			executableData, err = ioutil.ReadAll(tr)
			if err != nil {
				return err
			}
			break FOR
		case err == io.EOF:
			return fmt.Errorf("%s: could not find header", c.upgrade.repo)
		}
	}

	return mutator.WriteFile(executableFilename, executableData, 0755, nil)
}

func (c *Config) snapRefresh(fs vfs.FS) error {
	return c.run(fs, "", "snap", "refresh", c.upgrade.repo)
}

func (c *Config) upgradePackage(fs vfs.FS, mutator chezmoi.Mutator, rr *github.RepositoryRelease, useSudo bool) error {
	switch runtime.GOOS {
	case "darwin":
		return c.run(fs, "", "brew", "upgrade", c.upgrade.repo)
	case "linux":
		// Determine the package type and architecture.
		packageType, err := getPackageType(fs)
		if err != nil {
			return err
		}
		arch := runtime.GOARCH
		if archReplacement, ok := archReplacements[packageType]; ok {
			arch = archReplacement[arch]
		}

		// chezmoi does not build and distribute AUR packages, so instead rely
		// on pacman and the communnity package.
		if packageType == packageTypeAUR {
			var args []string
			if useSudo {
				args = append(args, "sudo")
			}
			args = append(args, "pacman", "-S", c.upgrade.repo)
			return c.run(fs, "", args[0], args[1:]...)
		}

		// Find the corresponding release asset.
		var releaseAsset *github.ReleaseAsset
		suffix := arch + "." + packageType
		for i, ra := range rr.Assets {
			if strings.HasSuffix(ra.GetName(), suffix) {
				releaseAsset = &rr.Assets[i]
				break
			}
		}
		if releaseAsset == nil {
			return fmt.Errorf("cannot find release asset (arch=%q, packageType=%q)", arch, packageType)
		}

		// Create a temporary directory for the package.
		var tempDir string
		if c.DryRun {
			tempDir = os.TempDir()
		} else {
			tempDir, err = ioutil.TempDir("", "chezmoi")
			if c.Verbose {
				fmt.Fprintf(c.Stdout(), "mkdir -p %s\n", tempDir)
			}
			if err != nil {
				return err
			}
			defer func() {
				_ = mutator.RemoveAll(tempDir)
			}()
		}

		data, err := c.downloadURL(releaseAsset.GetBrowserDownloadURL())
		if err != nil {
			return err
		}
		if err := c.verifyChecksum(rr, releaseAsset.GetName(), data); err != nil {
			return err
		}

		packageFilename := filepath.Join(tempDir, releaseAsset.GetName())
		if err := mutator.WriteFile(packageFilename, data, 0644, nil); err != nil {
			return err
		}

		// Install the package from disk.
		var args []string
		if useSudo {
			args = append(args, "sudo")
		}
		switch packageType {
		case packageTypeDEB:
			args = append(args, "dpkg", "-i", packageFilename)
		case packageTypeRPM:
			args = append(args, "rpm", "-U", packageFilename)
		}
		return c.run(fs, "", args[0], args[1:]...)
	default:
		return fmt.Errorf("%s: unsupported GOOS", runtime.GOOS)
	}
}

func (c *Config) verifyChecksum(rr *github.RepositoryRelease, name string, data []byte) error {
	checksums, err := c.getChecksums(rr)
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

func getMethod(fs vfs.Stater, executableFilename string) (string, error) {
	if ok, _ := vfs.Contains(fs, executableFilename, "/snap"); ok {
		return methodSnapRefresh, nil
	}
	info, err := fs.Stat(executableFilename)
	if err != nil {
		return "", err
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	executableInUserHomeDir, err := vfs.Contains(fs, executableFilename, userHomeDir)
	if err != nil {
		return "", err
	}
	executableIsInTempDir, err := vfs.Contains(fs, executableFilename, os.TempDir())
	if err != nil {
		return "", err
	}

	executableStat := info.Sys().(*syscall.Stat_t)
	uid := os.Getuid()
	switch runtime.GOOS {
	case "darwin":
		if int(executableStat.Uid) != uid {
			return "", fmt.Errorf("%s: cannot upgrade executable owned by non-current user", executableFilename)
		}
		if executableInUserHomeDir || executableIsInTempDir {
			return methodReplaceExecutable, nil
		}
		return methodUpgradePackage, nil
	case "freebsd":
		return methodReplaceExecutable, nil
	case "linux":
		if uid == 0 {
			if executableStat.Uid != 0 {
				return "", fmt.Errorf("%s: cannot upgrade executable owned by non-root user when running as root", executableFilename)
			}
			if executableInUserHomeDir || executableIsInTempDir {
				return methodReplaceExecutable, nil
			}
			return methodUpgradePackage, nil
		}
		switch int(executableStat.Uid) {
		case 0:
			method := methodUpgradePackage
			if _, err := exec.LookPath("sudo"); err == nil {
				method = methodSudoPrefix + method
			}
			return method, nil
		case uid:
			return methodReplaceExecutable, nil
		default:
			return "", fmt.Errorf("%s: cannot upgrade executable owned by non-current non-root user", executableFilename)
		}
	case "openbsd":
		return methodReplaceExecutable, nil
	default:
		return "", fmt.Errorf("%s: unsupported GOOS", runtime.GOOS)
	}
}

func getPackageType(fs vfs.FS) (string, error) {
	osRelease, err := getOSRelease(fs)
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
			return &rr.Assets[i]
		}
	}
	return nil
}
