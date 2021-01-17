package cmd

import (
	"context"
	"net/http"
	"os"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

type gitHubData struct {
	client    *github.Client
	keysCache map[string][]*github.Key
}

func (c *Config) gitHubKeysTemplateFunc(user string) []*github.Key {
	if keys, ok := c.gitHub.keysCache[user]; ok {
		return keys
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if c.gitHub.client == nil {
		var httpClient *http.Client
		for _, key := range []string{
			"CHEZMOI_GITHUB_ACCESS_TOKEN",
			"GITHUB_ACCESS_TOKEN",
			"GITHUB_TOKEN",
		} {
			if accessToken := os.Getenv(key); accessToken != "" {
				httpClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
					AccessToken: accessToken,
				}))
				break
			}
		}
		c.gitHub.client = github.NewClient(httpClient)
	}

	var allKeys []*github.Key
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		keys, resp, err := c.gitHub.client.Users.ListKeys(ctx, user, opts)
		if err != nil {
			returnTemplateError(err)
			return nil
		}
		allKeys = append(allKeys, keys...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if c.gitHub.keysCache == nil {
		c.gitHub.keysCache = make(map[string][]*github.Key)
	}
	c.gitHub.keysCache[user] = allKeys
	return allKeys
}
