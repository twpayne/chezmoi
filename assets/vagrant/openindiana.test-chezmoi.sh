#!/bin/bash

set -eufo pipefail

cd /chezmoi

go test ./...
