GO?=go
GOFUMPT_VERSION=$(shell awk '/GOFUMPT_VERSION:/ { print $$2 }' .github/workflows/main.yml)
GOLANGCI_LINT_VERSION=$(shell awk '/GOLANGCI_LINT_VERSION:/ { print $$2 }' .github/workflows/main.yml)
ifdef VERSION
	GO_LDFLAGS+=-X main.version=${VERSION}
endif
ifdef COMMIT
	GO_LDFLAGS+=-X main.commit=${COMMIT}
endif
ifdef DATE
	GO_LDFLAGS+=-X main.date=${DATE}
endif
ifdef BUILT_BY
	GO_LDFLAGS+=-X main.builtBy=${BUILT_BY}
endif
PREFIX?=/usr/local

.PHONY: default
default: build

.PHONY: smoketest
smoketest: run build-all test lint format

.PHONY: build
build:
ifeq (${GO_LDFLAGS},)
	go build . || ( rm -f chezmoi ; false )
else
	go build -ldflags "${GO_LDFLAGS}" . || ( rm -f chezmoi ; false )
endif

.PHONY: install
install: build
	install -m 755 chezmoi "${DESTDIR}${PREFIX}/bin"

.PHONY: install-from-git-working-copy
install-from-git-working-copy:
	go install -ldflags "-X main.version=$(shell git describe --abbrev=0 --tags) \
		-X main.commit=$(shell git rev-parse HEAD) \
		-X main.date=$(shell git show -s --format=%ct HEAD) \
		-X main.builtBy=source"

.PHONY: build-in-git-working-copy
build-in-git-working-copy:
	go build -ldflags "-X main.version=$(shell git describe --abbrev=0 --tags) \
		-X main.commit=$(shell git rev-parse HEAD) \
		-X main.date=$(shell git show -s --format=%ct HEAD) \
		-X main.builtBy=source"

.PHONY: build-all
build-all: build-darwin build-freebsd build-linux build-windows

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

.PHONY: test-all
test-all: test test-release rm-dist test-docker test-vagrant

.PHONY: rm-dist
rm-dist:
	rm -rf dist

.PHONY: test
test:
	${GO} test -ldflags="-X github.com/twpayne/chezmoi/pkg/chezmoitest.umaskStr=0o022" ./...
	${GO} test -ldflags="-X github.com/twpayne/chezmoi/pkg/chezmoitest.umaskStr=0o002" ./...

.PHONY: test-docker
test-docker:
	( cd assets/docker && ./test.sh alpine archlinux fedora voidlinux )

.PHONY: test-vagrant
test-vagrant:
	( cd assets/vagrant && ./test.sh debian11-i386 freebsd13 openindiana )

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
format: ensure-gofumpt
	find . -name \*.go | xargs ./bin/gofumpt -extra -w

.PHONY: ensure-tools
ensure-tools: ensure-gofumpt ensure-golangci-lint

.PHONY: ensure-gofumpt
ensure-gofumpt:
	if [ ! -x bin/gofumpt ] || ( ./bin/gofumpt --version | grep -Fqv "v${GOFUMPT_VERSION}" ) ; then \
		mkdir -p bin ; \
		GOBIN=$(shell pwd)/bin ${GO} install "mvdan.cc/gofumpt@v${GOFUMPT_VERSION}" ; \
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
