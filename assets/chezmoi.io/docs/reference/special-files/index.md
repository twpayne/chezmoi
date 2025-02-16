# Special files

All files in the source directory whose name begins with `.` are ignored by
default, unless they are one of the special files listed here. All of these
files are optional and are evaluated in a specific order.

- [`.chezmoiversion`][version] is read before any other file to ensure that the
  running version of chezmoi is new enough.

- [`.chezmoiroot`][root] is read after `.chezmoiversion` to determine which
  subdirectory is the root of the source state. All other special files are
  found in the source state determined by `.chezmoiroot`.

- [`.chezmoi.$FORMAT.tmpl`][config] is used by [`chezmoi init`][init] to prepare
  or update the chezmoi config file. This is also used when the command supports
  the `--init` flag, such as `chezmoi apply --init`. This will be applied
  _prior_ to any other special files.

- [`.chezmoidata.$FORMAT`][data] files are read before any templates are
  processed so that data contained within are available to the templates. See
  also [`.chezmoidata/` directories][data-dir].

- [`.chezmoiingore`][ignore] determines files that should be ignored.

- [`.chezmoiremove`][remove] determines files that should be removed during an
  apply.

- [`.chezmoiexternal.$FORMAT`][external] includes external files and archives as
  if they were in the source state. See also
  [`.chezmoiexternals/` directories][external-dir].

[config]: /reference/special-files/chezmoi-format-tmpl.md
[data]: /reference/special-files/chezmoidata-format.md
[data-dir]: /reference/special-directories/chezmoidata.md
[external]: /reference/special-files/chezmoiexternal-format.md
[external-dir]: /reference/special-directories/chezmoiexternals.md
[ignore]: /reference/special-files/chezmoiignore.md
[init]: /reference/commands/init.md
[remove]: /reference/special-files/chezmoiremove.md
[root]: /reference/special-files/chezmoiroot.md
[version]: /reference/special-files/chezmoiversion.md
