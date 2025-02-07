# `gitHubReleaseAssetURL` *owner-repo* *version* *pattern*

`gitHubReleaseAssetURL` calls the GitHub API to retrieve the latest
releases about the given *owner-repo*, returning structured data as defined by
the [GitHub Go API
bindings](https://pkg.go.dev/github.com/google/go-github/v69/github#RepositoryRelease).
It iterates through all the versions of the release, returning the first entry equal to *version*.
It then iterates through all the release's assets, returning the first one that
matches *pattern*. *pattern* is a shell pattern as [described in
`path.Match`](https://pkg.go.dev/path#Match).

Calls to `gitHubReleaseAssetURL` are cached so calling
`gitHubReleaseAssetURL` with the same *owner-repo* will only result in one
call to the GitHub API.

!!! example

    ```
    {{ gitHubReleaseAssetURL "FiloSottile/age" "age v1.2.0" (printf "age-*-%s-%s.tar.gz" .chezmoi.os .chezmoi.arch) }}
    {{ gitHubReleaseAssetURL "twpayne/chezmoi" "v2.50.0" (printf "chezmoi-%s-%s" .chezmoi.os .chezmoi.arch) }}
    ```
