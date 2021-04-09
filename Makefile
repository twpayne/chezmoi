GO?=go
GOLANGCI_LINT_VERSION=1.39.0

.PHONY: default
default: build run test lint format

.PHONY: build
build: build-darwin build-linux build-windows

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GO) build -o /dev/null .
	GOOS=darwin GOARCH=arm64 $(GO) build -o /dev/null .

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build -o /dev/null .
	GOOS=linux GOARCH=amd64 $(GO) build -tags=noupgrade -o /dev/null .

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build -o /dev/null .

.PHONY: run
run:
	$(GO) run . --version

.PHONY: test
test:
	$(GO) test -ldflags="-X github.com/twpayne/chezmoi/internal/chezmoitest.umaskStr=0o022" ./...
	$(GO) test -ldflags="-X github.com/twpayne/chezmoi/internal/chezmoitest.umaskStr=0o002" ./...

.PHONY: generate
generate: completions assets/scripts/install.sh

.PHONY: completions
completions:
	$(GO) run . completion bash -o completions/chezmoi-completion.bash
	$(GO) run . completion fish -o completions/chezmoi.fish
	$(GO) run . completion powershell -o completions/chezmoi.ps1
	$(GO) run . completion zsh -o completions/chezmoi.zsh

assets/scripts/install.sh: internal/cmd/generate-install.sh/install.sh.tmpl internal/cmd/generate-install.sh/main.go
	$(GO) run ./internal/cmd/generate-install.sh > $@

.PHONY: lint
lint: ensure-golangci-lint
	./bin/golangci-lint run
	$(GO) run ./internal/cmd/lint-whitespace

.PHONY: format
format: ensure-gofumports
	find . -name \*.go | xargs ./bin/gofumports -local github.com/twpayne/chezmoi -w

.PHONY: ensure-tools
ensure-tools: ensure-gofumports ensure-golangci-lint

.PHONY: ensure-gofumports
ensure-gofumports:
	if [ ! -x bin/gofumports ] ; then \
		mkdir -p bin ; \
		GOBIN=$(shell pwd)/bin $(GO) install mvdan.cc/gofumpt/gofumports@latest ; \
	fi

.PHONY: ensure-golangci-lint
ensure-golangci-lint:
	if [ ! -x bin/golangci-lint ] || ( ./bin/golangci-lint --version | grep -Fqv "version ${GOLANGCI_LINT_VERSION}" ) ; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- v${GOLANGCI_LINT_VERSION} ; \
	fi

.PHONY: release
release:
	goreleaser release \
		--rm-dist \
		${GORELEASER_FLAGS}

.PHONY: test-release
test-release:
	goreleaser release \
		--rm-dist \
		--skip-publish \
		--snapshot \
		${GORELEASER_FLAGS}

.PHONY: update-devcontainer
update-devcontainer:
	rm -rf .devcontainer && mkdir .devcontainer && curl -sfL https://github.com/microsoft/vscode-dev-containers/archive/master.tar.gz | tar -xzf - -C .devcontainer --strip-components=4 vscode-dev-containers-master/containers/go/.devcontainer
