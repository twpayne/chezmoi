package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/go-github/v64/github"
	"github.com/google/renameio/v2/maybe"
	"gopkg.in/yaml.v3"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

var (
	templateDataFilename = flag.String("data", "", "data filename")
	outputFilename       = flag.String("output", "", "output filename")
)

type gitHubClient struct {
	ctx    context.Context //nolint:containedctx
	client *github.Client
}

func newGitHubClient(ctx context.Context) *gitHubClient {
	return &gitHubClient{
		ctx:    ctx,
		client: chezmoi.NewGitHubClient(ctx, http.DefaultClient),
	}
}

func (c *gitHubClient) gitHubListReleases(ownerRepo string) []*github.RepositoryRelease {
	owner, repo, ok := strings.Cut(ownerRepo, "/")
	if !ok {
		panic(fmt.Errorf("%s: not a owner/repo", ownerRepo))
	}

	var allRepositoryReleases []*github.RepositoryRelease
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		repositoryReleases, resp, err := c.client.Repositories.ListReleases(c.ctx, owner, repo, opts)
		if err != nil {
			panic(err)
		}
		allRepositoryReleases = append(allRepositoryReleases, repositoryReleases...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allRepositoryReleases
}

func (c *gitHubClient) gitHubLatestRelease(ownerRepo string) *github.RepositoryRelease {
	owner, repo, ok := strings.Cut(ownerRepo, "/")
	if !ok {
		panic(fmt.Errorf("%s: not a owner/repo", ownerRepo))
	}

	rr, _, err := c.client.Repositories.GetLatestRelease(c.ctx, owner, repo)
	if err != nil {
		panic(err)
	}

	return rr
}

func run() error {
	flag.Parse()

	var templateData any
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
		return errors.New("no arguments")
	}

	templateName := path.Base(flag.Arg(0))
	buffer := &bytes.Buffer{}
	funcMap := sprig.TxtFuncMap()
	gitHubClient := newGitHubClient(context.Background())
	funcMap["exists"] = func(name string) bool {
		switch _, err := os.Stat(name); {
		case err == nil:
			return true
		case errors.Is(err, fs.ErrNotExist):
			return false
		default:
			panic(err)
		}
	}
	funcMap["gitHubLatestRelease"] = gitHubClient.gitHubLatestRelease
	funcMap["gitHubListReleases"] = gitHubClient.gitHubListReleases
	funcMap["gitHubTimestampFormat"] = func(layout string, timestamp github.Timestamp) string {
		return timestamp.Format(layout)
	}
	funcMap["output"] = func(name string, args ...string) string {
		cmd := exec.Command(name, args...)
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			panic(err)
		}
		return string(out)
	}
	funcMap["replaceAllRegex"] = func(expr, repl, s string) string {
		return regexp.MustCompile(expr).ReplaceAllString(s, repl)
	}
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
		if err := maybe.WriteFile(*outputFilename, buffer.Bytes(), 0o644); err != nil {
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
