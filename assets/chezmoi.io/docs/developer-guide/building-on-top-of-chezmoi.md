# Building on top of chezmoi

chezmoi is designed with UNIX-style composability in mind, and the command line
tool is semantically versioned. Building on top of chezmoi should primarily be
done by executing the binary with arguments and the standard input and output
configured appropriately. The `chezmoi dump` and `chezmoi state` commands
allows the inspection of chezmoi's internal state.
