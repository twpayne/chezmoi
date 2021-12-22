#!/bin/bash

set -eufo pipefail

cd /chezmoi

go test ./...

sh assets/scripts/install.sh
bin/chezmoi --version
