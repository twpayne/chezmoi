# `nativeLineEndings` *text*

`nativeLineEndings` returns *text* with all line endings replaced with native
line endings. On Windows systems this means that `\n` is replaced with `\r\n`
and on non-Windows systems `\r\n` is replaced with `\n`.
