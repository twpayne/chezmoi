package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/renameio/v2/maybe"
	"gopkg.in/yaml.v3"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

var (
	templateDataFilename = flag.String("data", "", "data filename")
	outputFilename       = flag.String("output", "", "output filename")
)

func gitHubLatestRelease(userRepo string) string {
	user, repo, ok := chezmoi.CutString(userRepo, "/")
	if !ok {
		panic(fmt.Errorf("%s: not a user/repo", userRepo))
	}

	ctx := context.Background()

	client := chezmoi.NewGitHubClient(ctx, http.DefaultClient)

	rr, _, err := client.Repositories.GetLatestRelease(ctx, user, repo)
	if err != nil {
		panic(err)
	}

	return strings.TrimPrefix(rr.GetName(), "v")
}

func run() error {
	flag.Parse()

	var templateData interface{}
	if *templateDataFilename != "" {
		dataBytes, err := os.ReadFile(*templateDataFilename)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(dataBytes, &templateData); err != nil {
			return err
		}
	}

	if flag.NArg() == 0 {
		return fmt.Errorf("no arguments")
	}

	templateName := path.Base(flag.Arg(0))
	buffer := &bytes.Buffer{}
	funcMap := sprig.TxtFuncMap()
	funcMap["gitHubLatestRelease"] = gitHubLatestRelease
	tmpl, err := template.New(templateName).Funcs(funcMap).ParseFiles(flag.Args()...)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(buffer, templateData); err != nil {
		return err
	}

	if *outputFilename == "" {
		if _, err := os.Stdout.Write(buffer.Bytes()); err != nil {
			return err
		}
	} else if data, err := os.ReadFile(*outputFilename); err != nil || !bytes.Equal(data, buffer.Bytes()) {
		if err := maybe.WriteFile(*outputFilename, buffer.Bytes(), 0o666); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
