#!/bin/bash

set -eufo pipefail

GO_VERSION=$(grep GO_VERSION: /chezmoi/.github/workflows/main.yml | awk '{ print $2 }' )

go get "golang.org/dl/go${GO_VERSION}"
"${HOME}/go/bin/go${GO_VERSION}" download
export PATH="${HOME}/sdk/go${GO_VERSION}/bin:${PATH}"

( cd /chezmoi && go test ./... )
