GOLANGCI_LINT_VERSION=1.35.2

.PHONY: default
default: generate build run test lint format

.PHONY: build
build: build-darwin build-linux build-windows

.PHONY: build-darwin
build-darwin: generate
	GOOS=darwin GOARCH=amd64 go build -o /dev/null .
	GOOS=darwin GOARCH=amd64 go build -o /dev/null ./chezmoi2

.PHONY: build-linux
build-linux: generate
	GOOS=linux GOARCH=amd64 go build -o /dev/null .
	GOOS=linux GOARCH=amd64 go build -o /dev/null ./chezmoi2

.PHONY: build-windows
build-windows: generate
	GOOS=windows GOARCH=amd64 go build -o /dev/null .
	GOOS=windows GOARCH=amd64 go build -o /dev/null ./chezmoi2

.PHONY: run
run: generate
	go run . --version
	go run ./chezmoi2 --version

.PHONY: generate
generate:
	go generate
	go generate ./chezmoi2

.PHONY: test
test: generate
	go test ./...

.PHONY: generate-install.sh
generate-install.sh:
	go run ./internal/cmd/generate-install.sh > assets/scripts/install.sh

.PHONY: lint
lint: ensure-golangci-lint generate
	./bin/golangci-lint run
	go run ./internal/cmd/lint-whitespace

.PHONY: format
format: ensure-gofumports generate
	find . -name \*.go | xargs ./bin/gofumports -local github.com/twpayne/chezmoi -w

.PHONY: ensure-tools
ensure-tools: ensure-gofumports ensure-golangci-lint

.PHONY: ensure-gofumports
ensure-gofumports:
	if [ ! -x bin/gofumports ] ; then \
		mkdir -p bin ; \
		( cd $$(mktemp -d) && go mod init tmp && GOBIN=$(shell pwd)/bin go get mvdan.cc/gofumpt/gofumports ) ; \
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
