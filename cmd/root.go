package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/absfs/afero"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile string
	config     Config
)

var rootCommand = &cobra.Command{
	Use:              "chezmoi",
	Short:            "chezmoi manages your home directory",
	PersistentPreRun: makeRun(persistentPreRunRoot),
}

func init() {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	persistentFlags := rootCommand.PersistentFlags()

	persistentFlags.StringVarP(&configFile, "config", "c", filepath.Join(homeDir, ".chezmoi.yaml"), "config file")

	persistentFlags.BoolVarP(&config.DryRun, "dry-run", "n", false, "dry run")
	viper.BindPFlag("dry-run", persistentFlags.Lookup("dry-run"))

	persistentFlags.StringVarP(&config.SourceDir, "source", "s", filepath.Join(homeDir, ".chezmoi"), "source directory")
	viper.BindPFlag("source", persistentFlags.Lookup("source"))

	persistentFlags.StringVarP(&config.TargetDir, "target", "t", homeDir, "target directory")
	viper.BindPFlag("target", persistentFlags.Lookup("target"))

	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", false, "verbose")
	viper.BindPFlag("verbose", persistentFlags.Lookup("verbose"))

	cobra.OnInitialize(func() {
		if _, err := os.Stat(configFile); !os.IsNotExist(err) {
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				log.Fatal(err)
			}
			if err := viper.Unmarshal(&config); err != nil {
				log.Fatal(err)
			}
		}
	})
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}

func persistentPreRunRoot(fs afero.Fs, command *cobra.Command, args []string) error {
	fi, err := fs.Stat(config.SourceDir)
	switch {
	case err == nil && !fi.Mode().IsDir():
		return errors.Errorf("%s: not a directory", config.SourceDir)
	case err == nil && fi.Mode()&os.ModePerm != 0700:
		log.Printf("%s: want permissions 0700, got 0%o", config.SourceDir, fi.Mode()&os.ModePerm)
	case os.IsNotExist(err):
	default:
		return err
	}
	return nil
}
