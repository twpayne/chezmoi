# Plugins

chezmoi supports plugins, similar to git.

If you run `chezmoi command` where *command* is not a builtin chezmoi command
then chezmoi will look for a binary called `chezmoi-command` in your `$PATH`. If
such a binary is found then chezmoi will execute it. Otherwise, chezmoi will
report an unknown command error.
