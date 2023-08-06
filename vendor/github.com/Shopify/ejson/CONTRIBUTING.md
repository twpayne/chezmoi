## Dependencies

Take a look at `dev.yml`, which enumerates the development dependencies. In most cases, you'll only
need a modern Go and a modern Ruby. This file also lists the major development tasks (building,
running tests).

## Building on FreeBSD

Building on FreeBSD requires the GNU tools: `bash`, `gmake` and `gtar`.

You can install them with `pkg install bash gmake gtar`. Then use `gmake` in place of `make`.

## Contributions

We're happy to accept bugfixes and other PRs, but be aware that we are very conservative about
adding new features to EJSON. There's nothing unusual about our workflow, just fork/branch/PR.
