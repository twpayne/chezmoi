#!/bin/bash

set -eufo pipefail

git config --global --add safe.directory /chezmoi

cd /chezmoi

go test -tags=test ./...

sh assets/scripts/install.sh
bin/chezmoi --version
