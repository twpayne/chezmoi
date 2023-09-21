# `gitHubTags` *owner-repo*

`gitHubTags` calls the GitHub API to retrieve the first page of tags for
the given *owner-repo*, returning structured data as defined by the [GitHub Go
API
bindings](https://pkg.go.dev/github.com/google/go-github/v55/github#RepositoryTag).

Calls to `gitHubTags` are cached so calling `gitHubTags` with the
same *owner-repo* will only result in one call to the GitHub API.

!!! example

    ```
    {{ (index (gitHubTags "docker/compose") 0).Name }}
    ```

!!! note

    The maximum number of items returned by `gitHubReleases` is determined by
    default page size for the GitHub API.

!!! warning

    The values returned by `gitHubTags` are not directly queryable via the
    [`jq`](/reference/templates/functions/jq/) function and must instead be
    converted through JSON:

    ```
    {{ gitHubTags "docker/compose" | toJson | fromJson | jq ".[0].name" }}
    ```
