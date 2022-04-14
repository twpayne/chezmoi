#!/bin/sh

set -euf

git config --global --add safe.directory /chezmoi

GO=${GO:-go}

cd /chezmoi
${GO} run . doctor || true
${GO} test ./...

sh assets/scripts/install.sh
bin/chezmoi --version
