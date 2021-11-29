#!/bin/sh

set -euf

cd /chezmoi
go run . doctor || true
go test ./...
