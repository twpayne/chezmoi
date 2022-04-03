# GitHub functions

The `gitHub*` template functions return data from the GitHub API.

By default, an anonymous GitHub API request will be made, which is subject to
[GitHub's rate
limits](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)
(currently 60 requests per hour per source IP address). If any of the
environment variables `$CHEZMOI_GITHUB_ACCESS_TOKEN`, `$GITHUB_ACCESS_TOKEN`, or
`$GITHUB_TOKEN` are found, then the first one found will be used to authenticate
the GitHub API request, with a higher rate limit (currently 5,000 requests per
hour per user).

In practice, GitHub API rate limits are high enough that you should rarely need
to set a token, unless you are sharing a source IP address with many other
GitHub users. If needed, the GitHub documentation describes how to [create a
personal access
token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).
