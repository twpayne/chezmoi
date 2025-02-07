# `includeTemplate` *filename* [*data*]

`includeTemplate` returns the result of executing the contents of *filename*
with the optional *data*. Relative paths are first searched for in
`.chezmoitemplates` and, if not found, are interpreted relative to the source
directory.

+++ 2.22.0
