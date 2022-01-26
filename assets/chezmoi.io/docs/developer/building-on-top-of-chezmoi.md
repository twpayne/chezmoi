# Building on top of chezmoi

chezmoi is designed with UNIX-style composibility in mind, and the command line
tool is semantically versioned. Building on top of chezmoi should primarily be
done by executing the binary with arguments and the standard input and output
configured appropriately. The `chezmoi dump` and `chezmoi state` commands
allows the inspection of chezmoi's internal state.

chezmoi's internal functionality is available as the Go module
`github.com/twpayne/chezmoi/v2`, however there are no guarantees whatsoever
about the API stability of this module. The semantic version applies to the
command line tool, and not to any Go APIs at any level.
