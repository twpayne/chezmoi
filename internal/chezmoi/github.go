package chezmoi

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// NewGitHubClient returns a new github.Client configured with an access token
// and a http client, if available.
func NewGitHubClient(ctx context.Context, httpClient *http.Client, host string) *github.Client {
	for _, key := range accessTokenEnvKeys(host) {
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

func accessTokenEnvKeys(host string) []string {
	var keys []string
	for _, chezmoiKey := range []string{"CHEZMOI", ""} {
		hostKeys := []string{makeHostKey(host)}
		if host == "github.com" {
			hostKeys = append(hostKeys, "GITHUB")
		}
		for _, hostKey := range hostKeys {
			for _, accessKey := range []string{"ACCESS", ""} {
				key := strings.Join(nonEmpty([]string{chezmoiKey, hostKey, accessKey, "TOKEN"}), "_")
				keys = append(keys, key)
			}
		}
	}
	return keys
}

func makeHostKey(host string) string {
	// FIXME split host on non-ASCII characters
	// FIXME convert everything to uppercase
	// FIXME join components with _
	return host
}

func nonEmpty[S []T, T comparable](s S) S {
	// FIXME use something from slices for this
	result := make([]T, 0, len(s))
	var zero T
	for _, e := range s {
		if e != zero {
			result = append(result, e)
		}
	}
	return result
}
