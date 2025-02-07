# `passRaw` *pass-name*

`passRaw` returns passwords stored in [pass][pass] using the pass CLI (`pass`).
*pass-name* is passed to `pass show $PASS_NAME` and the output is returned. The
output from `pass` is cached so calling `passRaw` multiple times with the same
*pass-name* will only invoke `pass` once.

[pass]: https://www.passwordstore.org/
