#!/bin/sh

set -euf

cd /chezmoi
go run . doctor || true
go test ./...

sh assets/scripts/install.sh
bin/chezmoi --version
