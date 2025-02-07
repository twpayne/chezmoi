# `gitHubLatestRelease` _owner-repo_

`gitHubLatestRelease` calls the GitHub API to retrieve the latest release about
the given _owner-repo_, returning structured data as defined by the
[GitHub Go API bindings][bindings].

Calls to `gitHubLatestRelease` are cached so calling `gitHubLatestRelease` with
the same _owner-repo_ will only result in one call to the GitHub API.

!!! example

    ```
    {{ (gitHubLatestRelease "docker/compose").TagName }}
    ```

[bindings]: https://pkg.go.dev/github.com/google/go-github/v69/github#RepositoryRelease
