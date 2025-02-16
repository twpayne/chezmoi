# `gitHubRelease` *owner-repo* *version*

`gitHubRelease` calls the GitHub API to retrieve the latest releases about
the given *owner-repo*, It iterates through all the versions of the release,
fetching the first entry equal to *version*.

It then returns structured data as defined by the [GitHub Go API bindings][bindings].

Calls to `gitHubRelease` are cached so calling `gitHubRelease` with
the same *owner-repo* *version* will only result in one call to the GitHub API.

!!! example

    ```
    {{ (gitHubRelease "docker/compose" "v2.29.1").TagName }}
    ```

[bindings]: https://pkg.go.dev/github.com/google/go-github/v69/github#RepositoryRelease
