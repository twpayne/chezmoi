# `.chezmoiignore{,.tmpl}`

If a file called `.chezmoiignore` (with an optional `.tmpl` extension) exists in
the source state then it is interpreted as a set of patterns to ignore. Patterns
are matched using [`doublestar.Match`][match] and match against the target path,
not the source path.

Patterns can be excluded by prefixing them with a `!` character. All excludes
take priority over all includes.

Comments in `.chezmoiignore` files are introduced with the `#` character and run
to the end of the line. If there is a `#` character introduced after the
beginning of the line, it must be preceded by whitespace to be recognized as
a comment and not part of the file.

`.chezmoiignore` is interpreted as a template, whether or not it has a `.tmpl`
extension. This allows different files to be ignored on different machines.

`.chezmoiignore` files in source state subdirectories apply only to that
subdirectory.

!!! example

    ``` title="~/.local/share/chezmoi/.chezmoiignore"
    README.md

    *.txt   # ignore *.txt in the target directory
    */*.txt # ignore *.txt in subdirectories of the target directory
            # but not in subdirectories of subdirectories;
            # so a/b/c.txt would *not* be ignored

    */*.org# # Ignore org-mode backup files that end with `#`

    backups/   # ignore the backups folder, but not its contents
    backups/** # ignore the contents of backups folder but not the folder itself

    {{- if ne .email "firstname.lastname@company.com" }}
    # Ignore .company-directory unless configured with a company email
    .company-directory # note that the pattern is not dot_company-directory
    {{- end }}

    {{- if ne .email "me@home.org" }}
    .personal-file
    {{- end }}
    ```

[match]: https://pkg.go.dev/github.com/bmatcuk/doublestar/v4#Match
