// summarize-capslock summarizes the output of capslock.
//
// See https://github.com/google/capslock.
//
//go:generate go tool gojsonstruct --output=capslock.go --type-name=CapsLock testdata/capslock.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

func run() error {
	outputFilename := flag.String("output", "", "output filename")
	flag.Parse()

	// Decode the capslock output JSON from stdin.
	var capsLock CapsLock
	if err := json.NewDecoder(os.Stdin).Decode(&capsLock); err != nil {
		return err
	}

	// Build a map of packages to unique capabilities.
	capabilitiesByPackageSets := make(map[string]map[string]struct{})
	for _, capabilityInfo := range capsLock.CapabilityInfo {
		for _, path := range capabilityInfo.Path {
			if _, ok := capabilitiesByPackageSets[path.Package]; ok {
				capabilitiesByPackageSets[path.Package][capabilityInfo.Capability] = struct{}{}
			} else {
				capabilitiesByPackageSets[path.Package] = map[string]struct{}{
					capabilityInfo.Capability: {},
				}
			}
		}
	}
	capabilitiesByPackageSlices := make(map[string][]string)
PACKAGE:
	for _package, capabilities := range capabilitiesByPackageSets {
		switch components := strings.Split(_package, "/"); {
		case !strings.Contains(components[0], "."):
			continue PACKAGE // Skip standard library packages.
		case len(components) >= 2 && slices.Compare(components[:2], []string{"chezmoi.io", "chezmoi"}) == 0:
			continue PACKAGE // Skip our packages.
		}
		capabilitiesSlice := slices.Collect(maps.Keys(capabilities))
		sort.Strings(capabilitiesSlice)
		capabilitiesByPackageSlices[_package] = capabilitiesSlice
	}

	// Prepare the output writer.
	var output io.Writer
	if *outputFilename == "" || *outputFilename == "-" {
		output = os.Stdout
	} else {
		outputFile, err := os.Create(*outputFilename)
		if err != nil {
			return err
		}
		defer outputFile.Close()
		output = outputFile
	}

	// Write the map of packages to unique capabilities as YAML.
	return yaml.NewEncoder(output).Encode(capabilitiesByPackageSlices)
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
