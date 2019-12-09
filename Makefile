.PHONY: smoketest
smoketest:
	go run . --version
	go test ./...
	./bin/golangci-lint run

all: completions generate

.PHONY: completions
completions: \
	completions/chezmoi-completion.bash \
	completions/chezmoi.fish \
	completions/chezmoi.zsh

.PHONY: completions/chezmoi-completion.bash
completions/chezmoi-completion.bash:
	mkdir -p $$(dirname $@) && go run . completion bash > $@ || ( rm -f $@ ; false )

.PHONY: completions/chezmoi.fish
completions/chezmoi.fish:
	mkdir -p $$(dirname $@) && go run . completion fish > $@ || ( rm -f $@ ; false )

.PHONY: completions/chezmoi.zsh
completions/chezmoi.zsh:
	mkdir -p $$(dirname $@) && go run . completion zsh > $@ || ( rm -f $@ ; false )

.PHONY: format
format:
	find . -name \*.go | xargs $$(go env GOPATH)/bin/gofumports -w

.PHONY: generate
generate:
	go generate ./...
	$$(go env GOPATH)/bin/packr2

.PHONY: install-tools
install-tools:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- v1.21.0
	GO111MODULE=off go get -u \
		github.com/gobuffalo/packr/v2/packr2 \
		mvdan.cc/gofumpt/gofumports

.PHONY: lint
lint:
	./bin/golangci-lint run

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

.PHONY: test
test:
	go test -race ./...

.PHONY: update-install.sh
update-install.sh:
	curl -sfL -o scripts/install.sh https://install.goreleaser.com/github.com/twpayne/chezmoi.sh
