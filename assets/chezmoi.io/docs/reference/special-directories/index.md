# Special directories

All directories in the source state whose name begins with `.` are ignored by
default, unless they are one of the special directories listed here. All of
these directories are optional and are evaluated in a specific order described
in [special files][special-files].

- The files in [`.chezmoidata/`][data-dir] directories are read in lexical order
  with any [`.chezmoidata.$FORMAT`][data] files in the source state.

- The files in [`.chezmoitemplates/`][templates] are made available for use in
  source templates.

- The files in [`.chezmoiscripts/`][scripts] are read, templated, and according
  to their phase attributes (`run_after_`, `run_before_`, etc.) and lexical
  ordering.

- Files in [`.chezmoiexternals/`][externals-dir] are read in lexical order with
  any [`.chezmoiexternal.$FORMAT`][external] files.

[data-dir]: /reference/special-directories/chezmoidata.md
[data]: /reference/special-files/chezmoidata-format.md
[external]: /reference/special-files/chezmoiexternal-format.md
[externals-dir]: /reference/special-directories/chezmoiexternals.md
[scripts]: /reference/special-directories/chezmoiscripts.md
[templates]: /reference/special-directories/chezmoitemplates.md
[special-files]: /reference/special-files/index.md
