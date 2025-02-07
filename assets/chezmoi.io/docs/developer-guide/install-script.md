# Install script

chezmoi generates the [install script][install] from a single source of truth.
You must run

```sh
go generate
```

if your change includes any of the following:

* Modifications to the install script template.

* Additions or modifications to the list of supported operating systems and
  architectures.

chezmoi's continuous integration verifies that all generated files are up to
date. Changes to generated files should be included in the commit that modifies
the source of truth.

[install]: https://github.com/twpayne/chezmoi/blob/master/assets/scripts/install.sh
