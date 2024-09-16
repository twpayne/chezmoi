# Website

The [website](https://chezmoi.io) is generated with [Material for
MkDocs](https://squidfunk.github.io/mkdocs-material/) from the contents of the
`assets/chezmoi.io/docs/` directory. It is hosted by [GitHub
pages](https://pages.github.com/) from the [`gh-pages`
branch](https://github.com/twpayne/chezmoi/tree/gh-pages).

To build the website locally, both Go 1.22 (or later) and Python 3.10 (or later)
must be installed.

Change into the website directory:

```console
$ cd assets/chezmoi.io
```

!!! note ""

    === "Default"

        Install the website dependencies:

        ```console
        $ pip3 install --user -r requirements.txt
        ```

    === "virtualenv (Recommended)"

        Create a virtualenv with:

        ```console
        $ python3 -m venv .venv
        ```

        and [activate it](https://docs.python.org/3/library/venv.html#how-venvs-work).

        Install the website dependencies:

        ```console
        $ pip3 install -r requirements.txt
        ```

Test the website locally by running:

```console
$ mkdocs serve
```

and visiting [http://127.0.0.1:8000/](http://127.0.0.1:8000/).

Deploy the website with:

```console
$ mkdocs gh-deploy
```
