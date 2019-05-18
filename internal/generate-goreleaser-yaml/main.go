package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"text/template"
)

var (
	hostArch = flag.String("host-arch", runtime.GOARCH, "host architecture")
	hostOS   = flag.String("host-os", runtime.GOOS, "host OS")
	name     = flag.String("name", "goreleaser.yaml.tmpl", "template name")
)

func run() error {
	flag.Parse()
	tmpl, err := template.ParseFiles(flag.Args()...)
	if err != nil {
		return err
	}
	// This is a horrible horrible hack. The template itself generates a YAML
	// file where some of the values are they themselves templates. We need to
	// pass variables used in these values through unchanged. So, replace every
	// variable with the template expression that generates it. Yuk!
	return tmpl.ExecuteTemplate(os.Stdout, *name, struct {
		HostArch    string
		HostOS      string
		Arch        string
		Commit      string
		Date        string
		Env         map[string]string
		Os          string
		ProjectName string
		Tag         string
		Version     string
	}{
		HostArch: *hostArch,
		HostOS:   *hostOS,
		Arch:     "{{ .Arch }}",
		Commit:   "{{ .Commit }}",
		Date:     "{{ .Date }}",
		Env: map[string]string{
			"TRAVIS_BUILD_NUMBER": "{{ .Env.TRAVIS_BUILD_NUMBER }}",
		},
		Os:          "{{ .Os }}",
		ProjectName: "{{ .ProjectName }}",
		Tag:         "{{ .Tag }}",
		Version:     "{{ .Version }}",
	})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
