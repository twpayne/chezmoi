# Website

The [website](https://chezmoi.io) is generated with [Material for
MkDocs](https://squidfunk.github.io/mkdocs-material/) from the contents of the
`assets/chezmoi.io/docs/` directory. It hosted by [GitHub pages](https://pages.github.com/) from
the [`gh-pages` branch](https://github.com/twpayne/chezmoi/tree/gh-pages).

Install Material for MkDocs and the required plugins with:

```console
$ pip install mkdocs-material mkdocs-redirects mkdocs-simple-hooks
```

Test the website locally by running:

```console
$ cd assets/chezmoi.io
$ mkdocs serve
```

and visiting [https://127.0.0.1:8000/](http://127.0.0.1:8000/).

Deploy the website with:

```console
$ mkdocs gh-deploy
```
