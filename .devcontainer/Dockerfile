ARG VARIANT=dev-1.17
FROM mcr.microsoft.com/vscode/devcontainers/go:${VARIANT}

# [Optional] Uncomment this section to install additional OS packages.
RUN export DEBIAN_FRONTEND=noninteractive \
    && apt-get update \
    && apt-get install --yes --no-install-recommends acl musl-tools \
    && rm -rf /var/lib/apt/lists/*

# [Optional] Uncomment the next line to use go get to install anything else you need
RUN go get -x mvdan.cc/gofumpt

RUN curl -fsLS https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin"

RUN curl -fsLS https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh -s -- -b "$(go env GOPATH)/bin"
