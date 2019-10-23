.PHONY: nothing
nothing:

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

.PHONY: coverage.out
coverage.out:
	go test -cover -covermode=count -coverprofile=cmd-coverage.out -coverpkg=github.com/twpayne/chezmoi/cmd,github.com/twpayne/chezmoi/internal/chezmoi ./cmd
	go test -cover -covermode=count -coverprofile=internal-chezmoi-coverage.out ./internal/chezmoi
	$$(go env GOPATH)/bin/gocovmerge cmd-coverage.out internal-chezmoi-coverage.out > $@ || ( rm -f $@ ; false )

.PHONY: format
format:
	find . -name \*.go | xargs $$(go env GOPATH)/bin/gofumports -w

.PHONY: generate
generate:
	go generate ./...
	$$(go env GOPATH)/bin/packr2

.PHONY: html-coverage
html-coverage:
	go tool cover -html=coverage.out

.PHONY: install-tools
install-tools:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- v1.20.0
	GO111MODULE=off go get -u \
		golang.org/x/tools/cmd/cover \
		github.com/gobuffalo/packr/v2/packr2 \
		github.com/mattn/goveralls \
		github.com/wadey/gocovmerge \
		mvdan.cc/gofumpt/gofumports

.PHONY: lint
lint:
	go vet ./...
	./bin/golangci-lint run

.PHONY: release
release:
	goreleaser release \
		--rm-dist \
		${GORELEASER_FLAGS}

.PHONY: release-setup-travis
release-setup-travis:
	sudo snap install goreleaser --classic
	sudo snap install snapcraft --classic
	openssl aes-256-cbc \
		-K $${encrypted_b4d86685c6fa_key} \
		-iv $${encrypted_b4d86685c6fa_iv} \
		-in goreleaser/snap.login.enc \
		-out goreleaser/snap.login \
		-d
	snapcraft login \
		--with goreleaser/snap.login

.PHONY: test-release
test-release:
	TRAVIS_BUILD_NUMBER=1 goreleaser release \
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
