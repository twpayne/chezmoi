package chezmoi

import (
	"context"
	"net/http"
	"os"

	"github.com/google/go-github/v71/github"
	"golang.org/x/oauth2"
)

// NewGitHubClient returns a new github.Client configured with an access token
// and a http client, if available.
func NewGitHubClient(ctx context.Context, httpClient *http.Client) *github.Client {
	for _, key := range []string{
		"CHEZMOI_GITHUB_ACCESS_TOKEN",
		"CHEZMOI_GITHUB_TOKEN",
		"GITHUB_ACCESS_TOKEN",
		"GITHUB_TOKEN",
	} {
		if accessToken := os.Getenv(key); accessToken != "" {
			httpClient = oauth2.NewClient(
				context.WithValue(ctx, oauth2.HTTPClient, httpClient),
				oauth2.StaticTokenSource(&oauth2.Token{
					AccessToken: accessToken,
				}))
			break
		}
	}
	return github.NewClient(httpClient)
}
