package main

import (
	"cmp"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"text/template"

	"github.com/goccy/go-yaml"

	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

var (
	binDir = flag.String("b", "bin", "binary directory")
	output = flag.String("o", "", "output")
)

type platform struct {
	GOOS   string
	GOARCH string
}

func newPlatform(goos, goarch string) platform {
	return platform{
		GOOS:   goos,
		GOARCH: goarch,
	}
}

func (p platform) String() string {
	return p.GOOS + "/" + p.GOARCH
}

type platformValue struct {
	platform
	CgoSupported bool
}

type platformSet map[platform]platformValue

func goToolDistList() (platformSet, error) {
	cmd := exec.Command("go", "tool", "dist", "list", "-json")
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var platformValues []platformValue
	if err := json.Unmarshal(data, &platformValues); err != nil {
		return nil, err
	}

	result := make(platformSet)
	for _, pv := range platformValues {
		result[pv.platform] = pv
	}
	return result, nil
}

func run() error {
	flag.Parse()

	// Read goreleaser config.
	data, err := os.ReadFile(".goreleaser.yaml")
	if err != nil {
		return err
	}
	var goreleaserConfig struct {
		Builds []struct {
			GOOS   []string   `yaml:"goos"`
			GOARCH []string   `yaml:"goarch"`
			Ignore []platform `yaml:"ignore"`
		} `yaml:"builds"`
	}
	if err := yaml.Unmarshal(data, &goreleaserConfig); err != nil {
		return err
	}

	// Read list of supported platforms.
	supportedPlatforms, err := goToolDistList()
	if err != nil {
		return err
	}

	// Exclude unsupported platforms.
	delete(supportedPlatforms, newPlatform("windows", "arm64"))

	// Build set of platforms.
	allPlatforms := chezmoiset.New[platform]()
	for _, build := range goreleaserConfig.Builds {
		buildPlatforms := chezmoiset.New[platform]()
		for _, goos := range build.GOOS {
			for _, goarch := range build.GOARCH {
				platform := newPlatform(goos, goarch)
				if _, ok := supportedPlatforms[platform]; ok {
					buildPlatforms.Add(platform)
				}
			}
		}
		buildPlatforms.Remove(build.Ignore...)
		allPlatforms.AddSet(buildPlatforms)
	}

	// Sort platforms.
	sortedPlatforms := allPlatforms.Elements()
	slices.SortFunc(sortedPlatforms, func(a, b platform) int {
		return cmp.Compare(a.String(), b.String())
	})

	// Generate install.sh.
	installShTemplate, err := template.ParseFiles("internal/cmds/generate-install.sh/install.sh.tmpl")
	if err != nil {
		return err
	}
	var outputFile *os.File
	if *output == "" || *output == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(*output)
		if err != nil {
			return err
		}
		defer outputFile.Close()
	}
	return installShTemplate.ExecuteTemplate(outputFile, "install.sh.tmpl", struct {
		BinDir    string
		Platforms []platform
	}{
		BinDir:    *binDir,
		Platforms: sortedPlatforms,
	})
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
