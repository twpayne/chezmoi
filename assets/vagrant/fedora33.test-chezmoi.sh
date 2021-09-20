#!/bin/sh

GO_VERSION=$(grep GO_VERSION: /chezmoi/.github/workflows/main.yml | awk '{ print $2 }')
go get "golang.org/dl/go${GO_VERSION}"
"${HOME}/go/bin/go${GO_VERSION}" download
( cd /chezmoi && "${HOME}/go/bin/go${GO_VERSION}" test ./... )
