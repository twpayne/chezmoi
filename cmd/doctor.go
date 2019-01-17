package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"text/tabwriter"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

// doctorCmd represents the doctor command
var doctorCommand = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "doctor",
	Short: "Check your system for potential problems",
	Long: `Check your system for potential problems. Such as:
* Does the configuration file exist
* Does the destination directory exist
* Does the source directory exist
* Is Hashicorp Vault installed?
* Is Bitwarden installed?
* Is LastPass installed?
`,
	RunE: makeRunE(config.runDoctorCommand),
}

func init() {
	rootCommand.AddCommand(doctorCommand)
}

func (c *Config) runDoctorCommand(fs vfs.FS, args []string) error {
	exitWithErr := false
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintf(w, "Arch:\t%s\n", runtime.GOARCH)
	fmt.Fprintf(w, "OS:\t%s\n", runtime.GOOS)

	for _, dir := range []struct {
		description    string
		path           string
		requireErrCode bool
	}{
		{description: "Configuration File", path: configFile, requireErrCode: false},
		{description: "Destination Directory", path: config.DestDir, requireErrCode: true},
		{description: "Source Directory", path: config.SourceDir, requireErrCode: true},
	} {
		info, err := fs.Stat(dir.path)
		if os.IsNotExist(err) {
			fmt.Fprintf(w, "%s:\tDoes Not Exist\t(%s)\n", dir.description, dir.path)
			if dir.requireErrCode {
				exitWithErr = true
			}
		} else {
			fmt.Fprintf(w, "%s:\tExists\t(%s %s)\n", dir.description, info.Mode().Perm().String(), dir.path)
		}
	}

	for _, binary := range []struct {
		description string
		name        string
	}{
		{description: "Bitwarden", name: config.Bitwarden.BW},
		{description: "Editor", name: config.getEditor()},
		{description: "LastPass", name: config.LastPass.Lpass},
		{description: "Pass", name: config.Pass.Pass},
		{description: "Vault", name: config.Vault.Vault},
		{description: "VCS", name: config.SourceVCSCommand},
	} {
		path, err := exec.LookPath(binary.name)
		if err == nil {
			info, err := fs.Stat(path)
			if !os.IsNotExist(err) {
				fmt.Fprintf(w, "%s:\tInstalled\t(%s %s)\n", binary.description, info.Mode().Perm().String(), path)
			} else {
				fmt.Fprintf(w, "%s:\tError\t(%s)\n", binary.description, err.Error())
			}
		} else {
			fmt.Fprintf(w, "%s:\tNot Found\t(looking for: %s)\n", binary.description, binary.name)
		}
	}
	w.Flush()
	if exitWithErr {
		os.Exit(1)
	}
	return nil
}
