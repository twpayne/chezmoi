# Doppler

chezmoi includes support for [Doppler][doppler] using the `doppler` CLI to
expose data through the `doppler` and `dopplerProjectJson` template functions.

Log in using:

```sh
doppler login
```

It is now possible to interact with the `doppler` CLI in two different, but
similar, ways. Both make use of the command `doppler secrets download --json
--no-file` behind the scenes but present a different experience.

The `doppler` function is used in the following way:

```text
{{ doppler "SECRET_NAME" "project name" "config" }}
```

All secrets from the specified project/config combination are cached for
subsequent access and will not requery the `doppler` CLI for another secret in
the same project/config. This caching mechanism enhances performance and reduces
unnecessary CLI calls.

The `dopplerProjectJson` presents the secrets as `json` structured data and is
used in the following way:

```text
{{ (dopplerProjectJson "project" "config").PASSWORD }}
```

Additionally one can set the default values for the project and config (aka
environment) in your config file, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[doppler]
    project = "my-project"
    config = "dev"
```

With these default values, you can omit them in the call to both `doppler` and
`dopplerProjectJson`, for example:

```text
{{ doppler "SECRET_NAME" }}
{{ dopplerProjectJson.SECRET_NAME }}
```

It is important to note that neither of the above parse any individual secret as
`json`. This can be achieved by using the `fromJson` function, for example:

```text
{{ (doppler "SECRET_NAME" | fromJson).created_by.email_address }}
{{ (dopplerProjectJson.SECRET_NAME | fromJson).created_by.email_address }}
```

Obviously the secret would have to be saved in `json` format for this to work as
expected.

[doppler]: https://www.doppler.com
