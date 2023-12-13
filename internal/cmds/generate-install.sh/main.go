package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"text/template"

	"gopkg.in/yaml.v3"
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
	data, err := exec.Command("go", "tool", "dist", "list", "-json").Output()
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
	allPlatforms := make(map[platform]struct{})
	for _, build := range goreleaserConfig.Builds {
		buildPlatforms := make(map[platform]struct{})
		for _, goos := range build.GOOS {
			for _, goarch := range build.GOARCH {
				platform := newPlatform(goos, goarch)
				if _, ok := supportedPlatforms[platform]; ok {
					buildPlatforms[platform] = struct{}{}
				}
			}
		}
		for _, ignore := range build.Ignore {
			delete(buildPlatforms, ignore)
		}
		for platform := range buildPlatforms {
			allPlatforms[platform] = struct{}{}
		}
	}

	// Sort platforms.
	sortedPlatforms := make([]platform, 0, len(allPlatforms))
	for platform := range allPlatforms {
		sortedPlatforms = append(sortedPlatforms, platform)
	}
	sort.Slice(sortedPlatforms, func(i, j int) bool {
		return sortedPlatforms[i].String() < sortedPlatforms[j].String()
	})

	// Generate install.sh.
	installShTemplate, err := template.ParseFiles(
		"internal/cmds/generate-install.sh/install.sh.tmpl",
	)
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
