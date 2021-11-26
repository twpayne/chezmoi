.PHONY: ;

GO?=go
GOLANGCI_LINT_VERSION=$(shell grep GOLANGCI_LINT_VERSION: .github/workflows/main.yml | awk '{ print $$2 }')
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

default: build

smoketest: run build-all test lint format

build:
ifeq (${GO_LDFLAGS},)
	go build . || ( rm -f chezmoi ; false )
else
	go build -ldflags "${GO_LDFLAGS}" . || ( rm -f chezmoi ; false )
endif

install: build
	install -m 755 chezmoi "${DESTDIR}${PREFIX}/bin"

install-from-git-working-copy:
	go install -ldflags "-X main.version=$(shell git describe --abbrev=0 --tags) \
		-X main.commit=$(shell git rev-parse HEAD) \
		-X main.date=$(shell git show -s --format=%ct HEAD) \
		-X main.builtBy=source"

build-in-git-working-copy:
	go build -ldflags "-X main.version=$(shell git describe --abbrev=0 --tags) \
		-X main.commit=$(shell git rev-parse HEAD) \
		-X main.date=$(shell git show -s --format=%ct HEAD) \
		-X main.builtBy=source"

build-all: build-darwin build-freebsd build-linux build-windows

build-darwin:
	GOOS=darwin GOARCH=amd64 ${GO} build -o /dev/null .
	GOOS=darwin GOARCH=arm64 ${GO} build -o /dev/null .

build-freebsd:
	GOOS=freebsd GOARCH=amd64 ${GO} build -o /dev/null .

build-linux:
	GOOS=linux GOARCH=amd64 ${GO} build -o /dev/null .
	GOOS=linux GOARCH=amd64 ${GO} build -tags=noupgrade -o /dev/null .

build-windows:
	GOOS=windows GOARCH=amd64 ${GO} build -o /dev/null .

run:
	${GO} run . --version

test:
	${GO} test -ldflags="-X github.com/twpayne/chezmoi/internal/chezmoitest.umaskStr=0o022" ./...
	${GO} test -ldflags="-X github.com/twpayne/chezmoi/internal/chezmoitest.umaskStr=0o002" ./...

test-docker:
	( cd assets/docker && ./test.sh archlinux fedora voidlinux )

test-vagrant:
	( cd assets/vagrant && ./test.sh debian11-i386 freebsd13 openbsd6 openindiana )

coverage-html: coverage
	${GO} tool cover -html=coverage.out

coverage:
	${GO} test -coverprofile=coverage.out -coverpkg=./... ./...

generate:
	${GO} generate

lint: ensure-golangci-lint
	./bin/golangci-lint run
	${GO} run ./internal/cmds/lint-whitespace

format: ensure-gofumpt
	find . -name \*.go | xargs ./bin/gofumpt -w

ensure-tools: ensure-gofumpt ensure-golangci-lint

ensure-gofumpt:
	if [ ! -x bin/gofumpt ] ; then \
		mkdir -p bin ; \
		GOBIN=$(shell pwd)/bin ${GO} install mvdan.cc/gofumpt@v0.2.0 ; \
	fi

ensure-golangci-lint:
	if [ ! -x bin/golangci-lint ] || ( ./bin/golangci-lint --version | grep -Fqv "version ${GOLANGCI_LINT_VERSION}" ) ; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- v${GOLANGCI_LINT_VERSION} ; \
	fi

release:
	goreleaser release \
		--rm-dist \
		${GORELEASER_FLAGS}

test-release:
	goreleaser release \
		--rm-dist \
		--skip-publish \
		--snapshot \
		${GORELEASER_FLAGS}

update-devcontainer:
	rm -rf .devcontainer && mkdir .devcontainer && curl -sfL https://github.com/microsoft/vscode-dev-containers/archive/master.tar.gz | tar -xzf - -C .devcontainer --strip-components=4 vscode-dev-containers-master/containers/go/.devcontainer
