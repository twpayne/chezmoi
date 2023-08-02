# `doppler` *key* [*project* [*config*]]

`doppler` returns the secret for the specified project and configuration
from [Doppler](https://www.doppler.com) using `doppler secrets download --json --no-file`.

If either of *project* or *config* are empty or
omitted, then chezmoi will use the value from the
`doppler.project` and
`doppler.config` config variables if they are set and not empty.

!!! example

    ```
    {{ doppler "SECRET_NAME" "project_name" "configuration_name" }}
    ```
