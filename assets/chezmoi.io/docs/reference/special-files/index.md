# Special files

All files in the source directory whose name begins with `.` are ignored by
default, unless they are one of the special files listed here. All of these
files are optional and are evaluated in a specific order.

1. [`.chezmoiroot`][root] is read from the root of the source directory before
   anything other file, setting the source state path. The location of all other
   files, except `.chezmoiversion`, is relative to the source state path.

2. [`.chezmoi.$FORMAT.tmpl`][config] is used by [`chezmoi init`][init] to
   prepare or update the chezmoi config file. This is also used when the command
   supports the `--init` flag, such as `chezmoi apply --init`. This will be
   applied _prior_ to any remaining special files or directories.

3. Data files ([`.chezmoidata.$FORMAT`][data] files or files in
   [`.chezmoidata/` directories][data-dir]) are read before any templates are
   processed so that data contained within are available to the templates.

4. [`.chezmoitemplates/`][templates-dir] directories are made available for use
   in source templates.

5. [`.chezmoiignore`][ignore] determines files and directories that should be
   ignored.

6. [`.chezmoiremove`][remove] determines files that should be removed during an
   apply.

7. External sources ([`.chezmoiexternal.$FORMAT`][external] or files in
   [`.chezmoiexternals/`][externals-dir]) are read in lexical order to include
   external files and archives as if they were in the source state.

8. [`.chezmoiversion`][version] is processed before any operation is applied, to
   ensure that the running version of chezmoi is new enough.

[config]: /reference/special-files/chezmoi-format-tmpl.md
[data-dir]: /reference/special-directories/chezmoidata.md
[data]: /reference/special-files/chezmoidata-format.md
[external-dir]: /reference/special-directories/chezmoiexternals.md
[external]: /reference/special-files/chezmoiexternal-format.md
[externals-dir]: /reference/special-directories/chezmoiexternals.md
[ignore]: /reference/special-files/chezmoiignore.md
[init]: /reference/commands/init.md
[remove]: /reference/special-files/chezmoiremove.md
[root]: /reference/special-files/chezmoiroot.md
[templates-dir]: /reference/special-directories/chezmoitemplates.md
[version]: /reference/special-files/chezmoiversion.md
