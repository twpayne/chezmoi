GO?=go
GOLANGCI_LINT_VERSION=$(shell grep GOLANGCI_LINT_VERSION: .github/workflows/main.yml | awk '{ print $$2 }')

.PHONY: default
default: run build test lint format

.PHONY: install
install:
	go install -ldflags "-X main.version=$(shell git describe --abbrev=0 --tags) \
		-X main.commit=$(shell git rev-parse HEAD) \
		-X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
		-X main.builtBy=source"

.PHONY: build
build: build-darwin build-freebsd build-linux build-windows

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 ${GO} build -o /dev/null .
	GOOS=darwin GOARCH=arm64 ${GO} build -o /dev/null .

.PHONY: build-freebsd
build-freebsd:
	GOOS=freebsd GOARCH=amd64 ${GO} build -o /dev/null .

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 ${GO} build -o /dev/null .
	GOOS=linux GOARCH=amd64 ${GO} build -tags=noupgrade -o /dev/null .

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 ${GO} build -o /dev/null .

.PHONY: run
run:
	${GO} run . --version

.PHONY: test
test:
	${GO} test -ldflags="-X github.com/twpayne/chezmoi/internal/chezmoitest.umaskStr=0o022" ./...
	${GO} test -ldflags="-X github.com/twpayne/chezmoi/internal/chezmoitest.umaskStr=0o002" ./...

.PHONY: test-docker
test-docker:
	( cd assets/docker && ./test.sh archlinux fedora voidlinux )

.PHONY: test-vagrant
test-vagrant:
	( cd assets/vagrant && ./test.sh debian11-i386 freebsd13 openbsd6 )

.PHONY: coverage-html
coverage-html: coverage
	${GO} tool cover -html=coverage.out

.PHONY: coverage
coverage:
	${GO} test -coverprofile=coverage.out -coverpkg=./... ./...

.PHONY: generate
generate:
	${GO} generate

.PHONY: lint
lint: ensure-golangci-lint
	./bin/golangci-lint run
	${GO} run ./internal/cmds/lint-whitespace

.PHONY: format
format: ensure-gofumports
	find . -name \*.go | xargs ./bin/gofumports -local github.com/twpayne/chezmoi -w

.PHONY: ensure-tools
ensure-tools: ensure-gofumports ensure-golangci-lint

.PHONY: ensure-gofumports
ensure-gofumports:
	if [ ! -x bin/gofumports ] ; then \
		mkdir -p bin ; \
		GOBIN=$(shell pwd)/bin ${GO} install mvdan.cc/gofumpt/gofumports@latest ; \
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
