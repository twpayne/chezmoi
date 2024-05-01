package cmd

// FIXME store per-host state in persistent state

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/v62/github"

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

type gitHubHostOwnerRepo struct {
	Host  string
	Owner string
	Repo  string
}

type gitHubClientResult struct {
	client *github.Client
	err    error
}

type gitHubData struct {
	clientsByHost      map[string]gitHubClientResult
	keysCache          map[string][]*github.Key
	latestReleaseCache map[gitHubHostOwnerRepo]*github.RepositoryRelease
	releasesCache      map[gitHubHostOwnerRepo][]*github.RepositoryRelease
	tagsCache          map[gitHubHostOwnerRepo][]*github.RepositoryTag
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

	gitHubClient, err := c.getGitHubClient(ctx, "github.com")
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

func (c *Config) gitHubLatestReleaseAssetURLTemplateFunc(hostOwnerRepo, pattern string) string {
	release, err := c.gitHubLatestRelease(hostOwnerRepo)
	if err != nil {
		panic(err)
	}
	for _, asset := range release.Assets {
		if asset.Name == nil {
			continue
		}
		switch ok, err := path.Match(pattern, *asset.Name); {
		case err != nil:
			panic(err)
		case ok:
			return *asset.BrowserDownloadURL
		}
	}
	return ""
}

func (c *Config) gitHubLatestRelease(hostOwnerRepo string) (*github.RepositoryRelease, error) {
	hor, err := gitHubSplitHostOwnerRepo(hostOwnerRepo)
	if err != nil {
		return nil, err
	}

	if release := c.gitHub.latestReleaseCache[hor]; release != nil {
		return release, nil
	}

	now := time.Now()
	gitHubLatestReleaseKey := []byte(hor.Owner + "/" + hor.Repo)
	if c.GitHub.RefreshPeriod != 0 {
		var gitHubLatestReleaseStateValue gitHubLatestReleaseState
		switch ok, err := chezmoi.PersistentStateGet(c.persistentState, gitHubLatestReleaseStateBucket, gitHubLatestReleaseKey, &gitHubLatestReleaseStateValue); {
		case err != nil:
			return nil, err
		case ok && now.Before(gitHubLatestReleaseStateValue.RequestedAt.Add(c.GitHub.RefreshPeriod)):
			return gitHubLatestReleaseStateValue.Release, nil
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitHubClient, err := c.getGitHubClient(ctx, hor.Host)
	if err != nil {
		return nil, err
	}

	release, _, err := gitHubClient.Repositories.GetLatestRelease(ctx, hor.Owner, hor.Repo)
	if err != nil {
		return nil, err
	}

	if err := chezmoi.PersistentStateSet(c.persistentState, gitHubLatestReleaseStateBucket, gitHubLatestReleaseKey, &gitHubLatestReleaseState{
		RequestedAt: now,
		Release:     release,
	}); err != nil {
		return nil, err
	}

	if c.gitHub.latestReleaseCache == nil {
		c.gitHub.latestReleaseCache = make(map[gitHubHostOwnerRepo]*github.RepositoryRelease)
	}
	c.gitHub.latestReleaseCache[hor] = release

	return release, nil
}

func (c *Config) gitHubLatestReleaseTemplateFunc(hostOwnerRepo string) *github.RepositoryRelease {
	release, err := c.gitHubLatestRelease(hostOwnerRepo)
	if err != nil {
		panic(err)
	}
	return release
}

func (c *Config) gitHubLatestTagTemplateFunc(hostOwnerRepo string) *github.RepositoryTag {
	tags, err := c.getGitHubTags(hostOwnerRepo)
	if err != nil {
		panic(err)
	}

	if len(tags) > 0 {
		return tags[0]
	}

	return nil
}

func (c *Config) gitHubReleasesTemplateFunc(hostOwnerRepo string) []*github.RepositoryRelease {
	hor, err := gitHubSplitHostOwnerRepo(hostOwnerRepo)
	if err != nil {
		panic(err)
	}

	if releases := c.gitHub.releasesCache[hor]; releases != nil {
		return releases
	}

	now := time.Now()
	gitHubReleasesKey := []byte(hor.Owner + "/" + hor.Repo)
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

	gitHubClient, err := c.getGitHubClient(ctx, hor.Host)
	if err != nil {
		panic(err)
	}

	releases, _, err := gitHubClient.Repositories.ListReleases(ctx, hor.Owner, hor.Repo, nil)
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
		c.gitHub.releasesCache = make(map[gitHubHostOwnerRepo][]*github.RepositoryRelease)
	}
	c.gitHub.releasesCache[hor] = releases

	return releases
}

func (c *Config) gitHubTagsTemplateFunc(hostOwnerRepo string) []*github.RepositoryTag {
	tags, err := c.getGitHubTags(hostOwnerRepo)
	if err != nil {
		panic(err)
	}

	return tags
}

func (c *Config) getGitHubTags(hostOwnerRepo string) ([]*github.RepositoryTag, error) {
	hor, err := gitHubSplitHostOwnerRepo(hostOwnerRepo)
	if err != nil {
		return nil, err
	}

	if tags := c.gitHub.tagsCache[hor]; tags != nil {
		return tags, nil
	}

	now := time.Now()
	gitHubTagsKey := []byte(hor.Owner + "/" + hor.Repo)
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

	gitHubClient, err := c.getGitHubClient(ctx, hor.Host)
	if err != nil {
		return nil, err
	}

	tags, _, err := gitHubClient.Repositories.ListTags(ctx, hor.Owner, hor.Repo, nil)
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
		c.gitHub.tagsCache = make(map[gitHubHostOwnerRepo][]*github.RepositoryTag)
	}
	c.gitHub.tagsCache[hor] = tags

	return tags, nil
}

func (c *Config) getGitHubClient(ctx context.Context, host string) (*github.Client, error) {
	if gitHubClientResult, ok := c.gitHub.clientsByHost[host]; ok {
		return gitHubClientResult.client, gitHubClientResult.err
	}

	httpClient, err := c.getHTTPClient()
	if err != nil {
		c.gitHub.clientsByHost[host] = gitHubClientResult{
			err: err,
		}
		return nil, err
	}

	client := chezmoi.NewGitHubClient(ctx, httpClient, host)
	c.gitHub.clientsByHost[host] = gitHubClientResult{
		client: client,
	}

	return client, nil
}

func gitHubSplitHostOwnerRepo(hostOwnerRepo string) (gitHubHostOwnerRepo, error) {
	switch components := strings.Split(hostOwnerRepo, "/"); len(components) {
	case 2:
		return gitHubHostOwnerRepo{
			Host:  "github.com",
			Owner: components[0],
			Repo:  components[1],
		}, nil
	case 3:
		return gitHubHostOwnerRepo{
			Host:  components[0],
			Owner: components[1],
			Repo:  components[2],
		}, nil
	default:
		return gitHubHostOwnerRepo{}, fmt.Errorf("%s: not a [host/]owner/repo", hostOwnerRepo)
	}
}
