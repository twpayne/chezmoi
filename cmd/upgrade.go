package cmd

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v25/github"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
	"golang.org/x/oauth2"
)

const (
	methodReplaceExecutable = "replace-executable"
	methodUpgradePackage    = "upgrade-package"
	methodSudoPrefix        = "sudo-"

	packageTypeNone = ""
	packageTypeDEB  = "deb"
	packageTypeRPM  = "rpm"
)

var (
	packageTypeByID = map[string]string{
		"amzn":     packageTypeRPM,
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

	// Use a Github API token, if set.
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
			fmt.Fprintf(c.Stdout(), "chezmoi: already at the latest version (%s)", Version)
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

	// Execute the new version.
	return syscall.Exec(executableFilename, []string{executableFilename, "--version"}, os.Environ())
}

func (c *Config) replaceExecutable(mutator chezmoi.Mutator, executableFilename string, releaseVersion *semver.Version, rr *github.RepositoryRelease) error {
	// Find the corresponding release asset.
	releaseAssetName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", c.upgrade.repo, releaseVersion, runtime.GOOS, runtime.GOARCH)
	var releaseAsset *github.ReleaseAsset
	for _, ra := range rr.Assets {
		if ra.GetName() == releaseAssetName {
			releaseAsset = &ra
			break
		}
	}
	if releaseAsset == nil {
		return fmt.Errorf("%s: cannot find release asset", releaseAssetName)
	}

	// Download the asset.
	resp, err := http.Get(releaseAsset.GetBrowserDownloadURL())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: got a non-200 OK response: %d %s", releaseAsset.GetBrowserDownloadURL(), resp.StatusCode, resp.Status)
	}

	// Extract the executable from the archive.
	gzipr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzipr.Close()
	tr := tar.NewReader(gzipr)
	var data []byte
FOR:
	for {
		h, err := tr.Next()
		switch {
		case err == nil && h.Name == c.upgrade.repo:
			data, err = ioutil.ReadAll(tr)
			if err != nil {
				return err
			}
			break FOR
		case err == io.EOF:
			return fmt.Errorf("%s: could not find header", c.upgrade.repo)
		}
	}

	// Replace the executable.
	return mutator.WriteFile(executableFilename, data, 0755, nil)
}

func (c *Config) upgradePackage(fs vfs.FS, mutator chezmoi.Mutator, rr *github.RepositoryRelease, useSudo bool) error {
	switch runtime.GOOS {
	case "darwin":
		return c.run("", "brew", "upgrade", c.upgrade.repo)
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

		// Find the corresponding release asset.
		var releaseAsset *github.ReleaseAsset
		suffix := arch + "." + packageType
		for _, ra := range rr.Assets {
			if strings.HasSuffix(ra.GetName(), suffix) {
				releaseAsset = &ra
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

		// Download the package.
		packageFilename := filepath.Join(tempDir, releaseAsset.GetName())
		if c.Verbose {
			fmt.Fprintf(c.Stdout(), "curl -o %s -s -L %s\n", packageFilename, releaseAsset.GetBrowserDownloadURL())
		}
		resp, err := http.Get(releaseAsset.GetBrowserDownloadURL())
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s: got a non-200 OK response: %d %s", releaseAsset.GetBrowserDownloadURL(), resp.StatusCode, resp.Status)
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := mutator.WriteFile(packageFilename, data, 0644, nil); err != nil {
			return err
		}

		// Install the package.
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
		return c.run("", args[0], args[1:]...)
	default:
		return fmt.Errorf("%s: unsupported GOOS", runtime.GOOS)
	}
}

func getMethod(fs vfs.FS, executableFilename string) (string, error) {
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
	executableStat := info.Sys().(*syscall.Stat_t)
	uid := os.Getuid()
	switch runtime.GOOS {
	case "darwin":
		if int(executableStat.Uid) != uid {
			return "", fmt.Errorf("%s: cannot upgrade executable owned by non-current user", executableFilename)
		}
		if executableInUserHomeDir {
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
			if executableInUserHomeDir {
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
