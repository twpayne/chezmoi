# `gopassCat` *secret*

`gopassCat` returns binary secrets stored in [gopass][gopass] using the gopass
CLI (`gopass`). *secret* is passed to `gopass cat $SECRET` and the output is
returned. The output from `gopassCat` is cached so calling `gopassCat` multiple
times with the same *secret* will only invoke `gopass` once.

[gopass]: https://www.gopass.pw/
