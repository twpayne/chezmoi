# 1.3.2

* Bump Go version and update dependencies

# 1.3.1

* Fix rubygems build for arm64.

# 1.2.2

* Bump various dependencies and rebuild releases with a modern Go version

# 1.2.1

* Bugfix: 1.2.0 introduced an issue in informational output formatting (no obvious security impact).
  This release simply fixes that bug.

# 1.2.0

* Moves error output from `stdout` to `stderr`.
* Various development hygiene changes; should be no user impact.

# 1.1.0

* Add `--key-from-stdin` flag, where a private key, assumed to match the file's public key, is read
  directly from stdin instead of looking up a match in the keydir.
