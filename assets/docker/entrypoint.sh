#!/bin/sh

set -euf

GO=${GO:-go}

cd /chezmoi
${GO} run . doctor || true
${GO} test ./...

sh assets/scripts/install.sh
bin/chezmoi --version
