# Website

The [website][website] is generated with [Material for MkDocs][material] from
the contents of the `assets/chezmoi.io/docs/` directory. It is hosted by
[GitHub pages][pages] from the [`gh-pages` branch][branch].

## Build

To build the website locally, Go 1.26 (or later) and [uv][uv] 0.5.0 (or later)
must be installed. Python 3.14 (or later) is required, but may be installed with
`uv`:

!!! note ""

    If Python 3.14 (or later) is not currently installed, install it with `uv`:

    ```sh
    uv python install 3.14
    ```

Install the dependencies (the `--frozen` is optional but recommended):

```sh
uv sync --frozen
```

## Preview

Preview the website locally by running:

```sh
uv run task serve-docs
```

and visiting [http://127.0.0.1:8000/][serve].

## Deploy

The website is automatically deployed when new releases are created, but manual
deployments can be triggered by maintainers with push access to the [`gh-pages`
branch][gh-pages] using:

```sh
uv run task deploy-docs
```

[website]: https://chezmoi.io
[material]: https://squidfunk.github.io/mkdocs-material/
[pages]: https://pages.github.com/
[branch]: https://github.com/twpayne/chezmoi/tree/gh-pages
[uv]: https://docs.astral.sh/uv/getting-started/installation/
[serve]: http://127.0.0.1:8000/
[gh-pages]: https://github.com/twpayne/chezmoi/tree/gh-pages
