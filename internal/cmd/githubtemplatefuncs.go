package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type gitHubConfig struct {
	RefreshPeriod time.Duration `json:"refreshPeriod" mapstructure:"refreshPeriod" yaml:"refreshPeriod"`
}

type gitHubKeysState struct {
	RequestedAt time.Time     `json:"requestedAt" yaml:"requestedAt"`
	Keys        []*github.Key `json:"keys"        yaml:"keys"`
}

type gitHubLatestReleaseState struct {
	RequestedAt time.Time                 `json:"requestedAt" yaml:"requestedAt"`
	Release     *github.RepositoryRelease `json:"release"     yaml:"release"`
}

type gitHubReleasesState struct {
	RequestedAt time.Time                   `json:"requestedAt" yaml:"requestedAt"`
	Releases    []*github.RepositoryRelease `json:"releases"    yaml:"releases"`
}

type gitHubTagsState struct {
	RequestedAt time.Time               `json:"requestedAt" yaml:"requestedAt"`
	Tags        []*github.RepositoryTag `json:"tags"        yaml:"tags"`
}

var (
	gitHubKeysStateBucket          = []byte("gitHubLatestKeysState")
	gitHubLatestReleaseStateBucket = []byte("gitHubLatestReleaseState")
	gitHubReleasesStateBucket      = []byte("gitHubReleasesState")
	gitHubTagsStateBucket          = []byte("gitHubTagsState")
)

type gitHubData struct {
	client             *github.Client
	clientErr          error
	keysCache          map[string][]*github.Key
	latestReleaseCache map[string]map[string]*github.RepositoryRelease
	releasesCache      map[string]map[string][]*github.RepositoryRelease
	tagsCache          map[string]map[string][]*github.RepositoryTag
}

func (c *Config) gitHubKeysTemplateFunc(user string) []*github.Key {
	if keys, ok := c.gitHub.keysCache[user]; ok {
		return keys
	}

	now := time.Now()
	gitHubKeysKey := []byte(user)
	if c.GitHub.RefreshPeriod != 0 {
		var gitHubKeysValue gitHubKeysState
		switch ok, err := chezmoi.PersistentStateGet(c.persistentState, gitHubKeysStateBucket, gitHubKeysKey, &gitHubKeysValue); {
		case err != nil:
			panic(err)
		case ok && now.Before(gitHubKeysValue.RequestedAt.Add(c.GitHub.RefreshPeriod)):
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
		switch ok, err := chezmoi.PersistentStateGet(c.persistentState, gitHubLatestReleaseStateBucket, gitHubLatestReleaseKey, &gitHubLatestReleaseStateValue); {
		case err != nil:
			panic(err)
		case ok && now.Before(gitHubLatestReleaseStateValue.RequestedAt.Add(c.GitHub.RefreshPeriod)):
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

	if err := chezmoi.PersistentStateSet(c.persistentState, gitHubLatestReleaseStateBucket, gitHubLatestReleaseKey, &gitHubLatestReleaseState{
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

func (c *Config) gitHubLatestTagTemplateFunc(ownerRepo string) *github.RepositoryTag {
	tags, err := c.getGitHubTags(ownerRepo)
	if err != nil {
		panic(err)
	}

	if len(tags) > 0 {
		return tags[0]
	}

	return nil
}

func (c *Config) gitHubReleasesTemplateFunc(ownerRepo string) []*github.RepositoryRelease {
	owner, repo, err := gitHubSplitOwnerRepo(ownerRepo)
	if err != nil {
		panic(err)
	}

	if releases := c.gitHub.releasesCache[owner][repo]; releases != nil {
		return releases
	}

	now := time.Now()
	gitHubReleasesKey := []byte(owner + "/" + repo)
	if c.GitHub.RefreshPeriod != 0 {
		var gitHubReleasesStateValue gitHubReleasesState
		switch ok, err := chezmoi.PersistentStateGet(c.persistentState, gitHubReleasesStateBucket, gitHubReleasesKey, &gitHubReleasesStateValue); {
		case err != nil:
			panic(err)
		case ok && now.Before(gitHubReleasesStateValue.RequestedAt.Add(c.GitHub.RefreshPeriod)):
			return gitHubReleasesStateValue.Releases
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitHubClient, err := c.getGitHubClient(ctx)
	if err != nil {
		panic(err)
	}

	releases, _, err := gitHubClient.Repositories.ListReleases(ctx, owner, repo, nil)
	if err != nil {
		panic(err)
	}

	if err := chezmoi.PersistentStateSet(c.persistentState, gitHubReleasesStateBucket, gitHubReleasesKey, &gitHubReleasesState{
		RequestedAt: now,
		Releases:    releases,
	}); err != nil {
		panic(err)
	}

	if c.gitHub.releasesCache == nil {
		c.gitHub.releasesCache = make(map[string]map[string][]*github.RepositoryRelease)
	}
	if c.gitHub.releasesCache[owner] == nil {
		c.gitHub.releasesCache[owner] = make(map[string][]*github.RepositoryRelease)
	}
	c.gitHub.releasesCache[owner][repo] = releases

	return releases
}

func (c *Config) gitHubTagsTemplateFunc(ownerRepo string) []*github.RepositoryTag {
	tags, err := c.getGitHubTags(ownerRepo)
	if err != nil {
		panic(err)
	}

	return tags
}

func (c *Config) getGitHubTags(ownerRepo string) ([]*github.RepositoryTag, error) {
	owner, repo, err := gitHubSplitOwnerRepo(ownerRepo)
	if err != nil {
		return nil, err
	}

	if tags := c.gitHub.tagsCache[owner][repo]; tags != nil {
		return tags, nil
	}

	now := time.Now()
	gitHubTagsKey := []byte(owner + "/" + repo)
	if c.GitHub.RefreshPeriod != 0 {
		var gitHubTagsStateValue gitHubTagsState
		switch ok, err := chezmoi.PersistentStateGet(c.persistentState, gitHubTagsStateBucket, gitHubTagsKey, &gitHubTagsStateValue); {
		case err != nil:
			return nil, err
		case ok && now.Before(gitHubTagsStateValue.RequestedAt.Add(c.GitHub.RefreshPeriod)):
			return gitHubTagsStateValue.Tags, nil
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitHubClient, err := c.getGitHubClient(ctx)
	if err != nil {
		return nil, err
	}

	tags, _, err := gitHubClient.Repositories.ListTags(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}

	if err := chezmoi.PersistentStateSet(c.persistentState, gitHubTagsStateBucket, gitHubTagsKey, &gitHubTagsState{
		RequestedAt: now,
		Tags:        tags,
	}); err != nil {
		return nil, err
	}

	if c.gitHub.tagsCache == nil {
		c.gitHub.tagsCache = make(map[string]map[string][]*github.RepositoryTag)
	}
	if c.gitHub.tagsCache[owner] == nil {
		c.gitHub.tagsCache[owner] = make(map[string][]*github.RepositoryTag)
	}
	c.gitHub.tagsCache[owner][repo] = tags

	return tags, nil
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

func gitHubSplitOwnerRepo(ownerRepo string) (string, string, error) {
	owner, repo, ok := strings.Cut(ownerRepo, "/")
	if !ok {
		return "", "", fmt.Errorf("%s: not an owner/repo", ownerRepo)
	}
	return owner, repo, nil
}
