# Website

The [website](https://chezmoi.io) is generated with
[Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) from the
contents of the `assets/chezmoi.io/docs/` directory. It is hosted by
[GitHub pages](https://pages.github.com/) from the
[`gh-pages` branch](https://github.com/twpayne/chezmoi/tree/gh-pages).

To build the website locally, Go 1.23 (or later) and
[uv](https://docs.astral.sh/uv/getting-started/installation/) 0.4.15 (or later)
must be installed. Python 3.10 (or later) is required, but may be installed with
`uv`:

!!! note ""

    If Python 3.10 (or later) is not currently installed, install it with `uv`:

    ```console
    $ uv python install 3.10
    ```

Install the dependencies (the `--frozen` is optional but recommended):

```console
$ uv sync --frozen
```

Test the website locally by running:

```console
$ uv run task serve-docs
```

and visiting [http://127.0.0.1:8000/](http://127.0.0.1:8000/).

## Maintainers

The website is automatically deployed when new releases are created, but manual
deployments can be triggered by maintainers with appropriate access using:

```console
$ uv run task mkdocs gh-deploy
```
