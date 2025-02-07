# `dopplerProjectJson` [*project* [*config*]]

`dopplerProjectJson` returns the secret for the specified project and
configuration from [Doppler][doppler] using `doppler secrets download --json
--no-file` as `json` structured data.

If either of *project* or *config* are empty or
omitted, then chezmoi will use the value from the
`doppler.project` and
`doppler.config` config variables if they are set and not empty.

!!! example

    ```
    {{ (dopplerProjectJson "project_name" "configuration_name").SECRET_NAME }}
    ```

[doppler]: https://www.doppler.com
