.PHONY: nothing
nothing:

all: .goreleaser.yaml

.PHONY: coverage.out
coverage.out:
	go test -cover -covermode=count -coverprofile=cmd-coverage.out -coverpkg=github.com/twpayne/chezmoi/cmd,github.com/twpayne/chezmoi/lib/chezmoi ./cmd
	go test -cover -covermode=count -coverprofile=lib-chezmoi-coverage.out ./lib/chezmoi
	gocovmerge cmd-coverage.out lib-chezmoi-coverage.out > $@ || ( rm -f $@ ; false )

.PHONY: format
format:
	find . -name \*.go | xargs gofumports -w

.PHONY: generate
generate:
	go generate ./...

.goreleaser.yaml: goreleaser/goreleaser.yaml.tmpl internal/generate-goreleaser-yaml/main.go
	go run ./internal/generate-goreleaser-yaml \
		-host-arch amd64 \
		-host-os linux \
		$< > $@ \
		|| ( rm -f $@ ; false )

goreleaser/goreleaser.host.yaml: goreleaser/goreleaser.yaml.tmpl internal/generate-goreleaser-yaml/main.go
	go run ./internal/generate-goreleaser-yaml \
		$< > $@ \
		|| ( rm -f $@ ; false )

.PHONY: html-coverage
html-coverage:
	go tool cover -html=coverage.out

.PHONY: install-tools
install-tools:
	GO111MODULE=off go get -u \
		golang.org/x/tools/cmd/cover \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/mattn/goveralls \
		github.com/wadey/gocovmerge \
		mvdan.cc/gofumpt \
		mvdan.cc/gofumpt/gofumports

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run

.PHONY: release
release:
	goreleaser release \
		--rm-dist \
		${GORELEASER_FLAGS}

.PHONY: release-snap
release-snap:
	goreleaser release \
		--config=goreleaser/goreleaser.snap.yaml \
		--rm-dist \
		--skip-publish \
		${GORELEASER_FLAGS}
	for snap in dist/*.snap ; do \
		snapcraft push --release=stable $${snap} ; \
	done

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
test-release: goreleaser/goreleaser.host.yaml
	TRAVIS_BUILD_NUMBER=1 goreleaser release \
		--config goreleaser/goreleaser.host.yaml \
		--rm-dist \
		--skip-publish \
		--snapshot \
		${GORELEASER_FLAGS}

.PHONY: test-release-snap
test-release-snap:
	TRAVIS_BUILD_NUMBER=1 goreleaser release \
		--config goreleaser/goreleaser.snap.yaml \
		--rm-dist \
		--skip-publish \
		--snapshot \
		${GORELEASER_FLAGS}

.PHONY: test
test:
	go test -race ./...

.PHONY: update-install.sh
update-install.sh:
	# FIXME re-enable this when https://github.com/goreleaser/godownloader/pull/114 is merged
	#curl -sfL -o scripts/install.sh https://install.goreleaser.com/github.com/twpayne/chezmoi.sh
