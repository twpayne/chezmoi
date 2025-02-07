# `gitHubTags` *owner-repo*

`gitHubTags` calls the GitHub API to retrieve the first page of tags for the
given *owner-repo*, returning structured data as defined by the
[GitHub Go API bindings][github-go].

Calls to `gitHubTags` are cached so calling `gitHubTags` with the same
*owner-repo* will only result in one call to the GitHub API.

!!! example

    ```
    {{ (index (gitHubTags "docker/compose") 0).Name }}
    ```

!!! note

    The maximum number of items returned by `gitHubReleases` is determined by
    default page size for the GitHub API.

!!! warning

    The values returned by `gitHubTags` are not directly queryable via the
    [`jq`][jq] function and must instead be converted through JSON:

    ```
    {{ gitHubTags "docker/compose" | toJson | fromJson | jq ".[0].name" }}
    ```

[github-go]: https://pkg.go.dev/github.com/google/go-github/v69/github#RepositoryTag
[jq]: /reference/templates/functions/jq.md
