package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v51/github"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type gitHubConfig struct {
	RefreshPeriod time.Duration `json:"refreshPeriod" mapstructure:"refreshPeriod" yaml:"refreshPeriod"`
}

type gitHubKeysState struct {
	RequestedAt time.Time     `json:"requestedAt" yaml:"requestedAt"`
	Keys        []*github.Key `json:"keys" yaml:"keys"`
}

type gitHubLatestReleaseState struct {
	RequestedAt time.Time                 `json:"requestedAt" yaml:"requestedAt"`
	Release     *github.RepositoryRelease `json:"release" yaml:"release"`
}

type gitHubLatestTagState struct {
	RequestedAt time.Time             `json:"requestedAt" yaml:"requestedAt"`
	Tag         *github.RepositoryTag `json:"tag" yaml:"tag"`
}

var (
	gitHubKeysStateBucket          = []byte("gitHubLatestKeysState")
	gitHubLatestReleaseStateBucket = []byte("gitHubLatestReleaseState")
	gitHubLatestTagStateBucket     = []byte("gitHubLatestTagState")
)

type gitHubData struct {
	client             *github.Client
	clientErr          error
	keysCache          map[string][]*github.Key
	latestReleaseCache map[string]map[string]*github.RepositoryRelease
	latestTagCache     map[string]map[string]*github.RepositoryTag
}

func (c *Config) gitHubKeysTemplateFunc(user string) []*github.Key {
	if keys, ok := c.gitHub.keysCache[user]; ok {
		return keys
	}

	now := time.Now()
	gitHubKeysKey := []byte(user)
	if c.GitHub.RefreshPeriod != 0 {
		var gitHubKeysValue gitHubKeysState
		switch ok, err := chezmoi.PersistentStateGet(c.persistentState, gitHubKeysStateBucket, gitHubKeysKey, &gitHubKeysValue); { //nolint:lll
		case err != nil:
			panic(err)
		case ok && !now.After(gitHubKeysValue.RequestedAt.Add(c.GitHub.RefreshPeriod)):
			return gitHubKeysValue.Keys
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitHubClient, err := c.getGitHubClient(ctx)
	if err != nil {
		panic(err)
	}

	var allKeys []*github.Key
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		keys, resp, err := gitHubClient.Users.ListKeys(ctx, user, opts)
		if err != nil {
			panic(err)
		}
		allKeys = append(allKeys, keys...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if err := chezmoi.PersistentStateSet(c.persistentState, gitHubKeysStateBucket, gitHubKeysKey, &gitHubKeysState{
		RequestedAt: now,
		Keys:        allKeys,
	}); err != nil {
		panic(err)
	}

	if c.gitHub.keysCache == nil {
		c.gitHub.keysCache = make(map[string][]*github.Key)
	}
	c.gitHub.keysCache[user] = allKeys

	return allKeys
}

func (c *Config) gitHubLatestReleaseTemplateFunc(ownerRepo string) *github.RepositoryRelease {
	owner, repo, err := gitHubSplitOwnerRepo(ownerRepo)
	if err != nil {
		panic(err)
	}

	if release := c.gitHub.latestReleaseCache[owner][repo]; release != nil {
		return release
	}

	now := time.Now()
	gitHubLatestReleaseKey := []byte(owner + "/" + repo)
	if c.GitHub.RefreshPeriod != 0 {
		var gitHubLatestReleaseStateValue gitHubLatestReleaseState
		switch ok, err := chezmoi.PersistentStateGet(c.persistentState, gitHubLatestReleaseStateBucket, gitHubLatestReleaseKey, &gitHubLatestReleaseStateValue); { //nolint:lll
		case err != nil:
			panic(err)
		case ok && !now.After(gitHubLatestReleaseStateValue.RequestedAt.Add(c.GitHub.RefreshPeriod)):
			return gitHubLatestReleaseStateValue.Release
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitHubClient, err := c.getGitHubClient(ctx)
	if err != nil {
		panic(err)
	}

	release, _, err := gitHubClient.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		panic(err)
	}

	if err := chezmoi.PersistentStateSet(c.persistentState, gitHubLatestReleaseStateBucket, gitHubLatestReleaseKey, &gitHubLatestReleaseState{ //nolint:lll
		RequestedAt: now,
		Release:     release,
	}); err != nil {
		panic(err)
	}

	if c.gitHub.latestReleaseCache == nil {
		c.gitHub.latestReleaseCache = make(map[string]map[string]*github.RepositoryRelease)
	}
	if c.gitHub.latestReleaseCache[owner] == nil {
		c.gitHub.latestReleaseCache[owner] = make(map[string]*github.RepositoryRelease)
	}
	c.gitHub.latestReleaseCache[owner][repo] = release

	return release
}

func (c *Config) gitHubLatestTagTemplateFunc(userRepo string) *github.RepositoryTag {
	owner, repo, err := gitHubSplitOwnerRepo(userRepo)
	if err != nil {
		panic(err)
	}

	if tag, ok := c.gitHub.latestTagCache[owner][repo]; ok {
		return tag
	}

	now := time.Now()
	gitHubLatestTagKey := []byte(owner + "/" + repo)
	if c.GitHub.RefreshPeriod != 0 {
		var gitHubLatestTagValue gitHubLatestTagState
		switch ok, err := chezmoi.PersistentStateGet(c.persistentState, gitHubLatestTagStateBucket, gitHubLatestTagKey, &gitHubLatestTagValue); { //nolint:lll
		case err != nil:
			panic(err)
		case ok && !now.After(gitHubLatestTagValue.RequestedAt.Add(c.GitHub.RefreshPeriod)):
			return gitHubLatestTagValue.Tag
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitHubClient, err := c.getGitHubClient(ctx)
	if err != nil {
		panic(err)
	}

	tags, _, err := gitHubClient.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{
		PerPage: 1,
	})
	if err != nil {
		panic(err)
	}
	var tag *github.RepositoryTag
	if len(tags) > 0 {
		tag = tags[0]
	}

	if err := chezmoi.PersistentStateSet(c.persistentState, gitHubLatestTagStateBucket, gitHubLatestTagKey, &gitHubLatestTagState{ //nolint:lll
		RequestedAt: now,
		Tag:         tag,
	}); err != nil {
		panic(err)
	}

	if c.gitHub.latestTagCache == nil {
		c.gitHub.latestTagCache = make(map[string]map[string]*github.RepositoryTag)
	}
	if c.gitHub.latestTagCache[owner] == nil {
		c.gitHub.latestTagCache[owner] = make(map[string]*github.RepositoryTag)
	}
	c.gitHub.latestTagCache[owner][repo] = tag

	return tag
}

func (c *Config) getGitHubClient(ctx context.Context) (*github.Client, error) {
	if c.gitHub.client != nil || c.gitHub.clientErr != nil {
		return c.gitHub.client, c.gitHub.clientErr
	}

	httpClient, err := c.getHTTPClient()
	if err != nil {
		c.gitHub.clientErr = err
		return nil, err
	}

	c.gitHub.client = chezmoi.NewGitHubClient(ctx, httpClient)
	return c.gitHub.client, nil
}

func gitHubSplitOwnerRepo(userRepo string) (string, string, error) {
	user, repo, ok := strings.Cut(userRepo, "/")
	if !ok {
		return "", "", fmt.Errorf("%s: not a user/repo", userRepo)
	}
	return user, repo, nil
}
