# `passageRaw` *pass-name*

`passageRaw` returns passwords stored in
[passage](https://github.com/FiloSottile/passage) using the pass CLI
(`passage`). *pass-name* is passed to `passage show $PASS_NAME` and the output
is returned. The output from `passage` is cached so calling `passageRaw`
multiple times with the same *pass-name* will only invoke `passage` once.
