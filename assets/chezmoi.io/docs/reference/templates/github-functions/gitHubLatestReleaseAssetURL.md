# `gitHubLatestReleaseAssetURL` *owner-repo* *pattern*

`gitHubLatestReleaseAssetURL` calls the GitHub API to retrieve the latest
release about the given *owner-repo*, returning structured data as defined by
the [GitHub Go API
bindings](https://pkg.go.dev/github.com/google/go-github/v63/github#RepositoryRelease).
It then iterates through all the release's assets, returning the first one that
matches *pattern*. *pattern* is a shell pattern as [described in
`path.Match`](https://pkg.go.dev/path#Match).

Calls to `gitHubLatestReleaseAssetURL` are cached so calling
`gitHubLatestReleaseAssetURL` with the same *owner-repo* will only result in one
call to the GitHub API.

!!! example

    ```
    {{ gitHubLatestReleaseAssetURL "FiloSottile/age" (printf "age-*-%s-%s.tar.gz" .chezmoi.os .chezmoi.arch) }}
    {{ gitHubLatestReleaseAssetURL "twpayne/chezmoi" (printf "chezmoi-%s-%s" .chezmoi.os .chezmoi.arch) }}
    ```
