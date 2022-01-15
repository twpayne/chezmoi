# `gopassRaw` *gopass-name*

`gopass` returns raw passwords stored in [gopass](https://www.gopass.pw/) using
the gopass CLI (`gopass`). *gopass-name* is passed to `gopass show --noparsing
<gopass-name>` and the output is returned. The output from `gopassRaw` is
cached so calling `gopassRaw` multiple times with the same *gopass-name* will
only invoke `gopass` once.
