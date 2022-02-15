# `gitHubKeys` *user*

`gitHubKeys` returns *user*'s public SSH keys from GitHub using the GitHub API.
The returned value is a slice of structs with `.ID` and `.Key` fields.

!!! warning

    If you use this function to populate your `~/.ssh/authorized_keys` file
    then you potentially open SSH access to anyone who is able to modify or add
    to your GitHub public SSH keys, possibly including certain GitHub
    employees. You should not use this function on publicly-accessible machines
    and should always verify that no unwanted keys have been added, for example
    by using the `-v` / `--verbose` option when running `chezmoi apply` or
    `chezmoi update`.

By default, an anonymous GitHub API request will be made, which is subject to
[GitHub's rate
limits](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)
(currently 60 requests per hour per source IP address). If any of the
environment variables `$CHEZMOI_GITHUB_ACCESS_TOKEN`, `$GITHUB_ACCESS_TOKEN`,
or `$GITHUB_TOKEN` are found, then the first one found will be used to
authenticate the GitHub API request, with a higher rate limit (currently 5,000
requests per hour per user).

In practice, GitHub API rate limits are high enough that you should rarely need
to set a token, unless you are sharing a source IP address with many other
GitHub users. If needed, the GitHub documentation describes how to [create a
personal access
token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).

!!! example

    ```
    {{ range gitHubKeys "user" }}
    {{- .Key }}
    {{ end }}
    ```
