package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/google/go-github/v40/github"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

// gitHubAPIStateBucket is the bucket for recording the state of gitHub api response and ETag header.
var gitHubAPIStateBucket = []byte("gitHubApiState")

type gitHubData struct {
	keysCache map[string][]*github.Key
}

type cachedResponse struct {
	ETag string `json:"eTag"`
	Body []byte `json:"body"`
}

func (c *Config) gitHubKeysTemplateFunc(user string) []*github.Key {
	if keys, ok := c.gitHub.keysCache[user]; ok {
		return keys
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitHubClient := c.newConditionalRequestsGitHubClient(ctx)

	var allKeys []*github.Key
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		keys, resp, err := gitHubClient.Users.ListKeys(ctx, user, opts)
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

// newConditionalRequestsGitHubClient return a new github.Client with
// conditional requests functionality.
func (c *Config) newConditionalRequestsGitHubClient(ctx context.Context) *github.Client {
	persistentStateFileAbsPath, err := c.persistentStateFile()
	if err != nil {
		returnTemplateError(err)
		return nil
	}
	persistentState, err := chezmoi.NewBoltPersistentState(
		c.baseSystem, persistentStateFileAbsPath,
		chezmoi.BoltPersistentStateReadWrite,
	)
	if err != nil {
		returnTemplateError(err)
		return nil
	}

	httpClient := &http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet && req.Method != http.MethodHead {
				return http.DefaultTransport.RoundTrip(req)
			}

			key := []byte(req.Method + " " + req.URL.String())

			var cachedResp cachedResponse
			switch state, err := persistentState.Get(gitHubAPIStateBucket, key); {
			case err != nil:
				return nil, err
			case state != nil:
				if err = chezmoi.FormatJSON.Unmarshal(state, &cachedResp); err != nil {
					return nil, err
				}
				req.Header.Add("If-None-Match", cachedResp.ETag)
			}

			resp, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				return nil, err
			}

			switch resp.StatusCode {
			case http.StatusNotModified:
				resp.Body = io.NopCloser(bytes.NewBuffer(cachedResp.Body))
				// to pass the github.CheckResponse
				resp.StatusCode = http.StatusOK
			default:
				if eTag := resp.Header.Get("ETag"); eTag != "" {
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						return nil, err
					}
					state, err := chezmoi.FormatJSON.Marshal(cachedResponse{
						ETag: eTag,
						Body: bodyBytes,
					})
					if err != nil {
						return nil, err
					}
					if err := persistentState.Set(gitHubAPIStateBucket, key, state); err != nil {
						return nil, err
					}
					resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			return resp, err
		}),
	}

	gitHubClient := newGitHubClient(ctx, httpClient)
	return gitHubClient
}
