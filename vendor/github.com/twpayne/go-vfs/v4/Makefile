GO?=go
GOLANGCI_LINT_VERSION=$(shell grep GOLANGCI_LINT_VERSION: .github/workflows/main.yml | awk '{ print $$2 }')

.PHONY: smoketest
smoketest: test lint

.PHONY: test
test:
	${GO} test ./...

.PHONY: lint
lint: ensure-golangci-lint
	./bin/golangci-lint run

.PHONY: format
format: ensure-gofumports
	find . -name \*.go | xargs ./bin/gofumports -local github.com/twpayne/chezmoi -w

.PHONY: ensure-tools
ensure-tools: \
	ensure-gofumports \
	ensure-golangci-lint

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