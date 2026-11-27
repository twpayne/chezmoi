# Install script

chezmoi generates its [install scripts][install-scripts] from a
[template][generate-install].

If your change includes modifications to the install script template, or
additions or modifications to the list of supported operating systems and
architectures then you must run:

```sh
make generate
```

to re-generate the install scripts.

chezmoi's continuous integration verifies that all generated files are up to
date. Changes to generated files should be included in the commit that modifies
the source of truth.

Install scripts are deployed to the [website][website] when a new version of
chezmoi is [released][releases].

[install-scripts]: https://github.com/twpayne/chezmoi/blob/master/assets/scripts
[generate-install]: https://github.com/twpayne/chezmoi/tree/master/internal/cmds/generate-install.sh
[releases]: /developer-guide/releases.md
[website]: /developer-guide/website.md
