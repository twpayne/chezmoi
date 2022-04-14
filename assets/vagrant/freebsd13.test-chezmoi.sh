#!/bin/bash

set -eufo pipefail

git config --global --add safe.directory /chezmoi

cd /chezmoi

go test ./...

sh assets/scripts/install.sh
bin/chezmoi --version
