NAME=ejson
RUBY_MODULE=EJSON
PACKAGE=github.com/Shopify/ejson
VERSION=$(shell cat VERSION)
GEM=pkg/$(NAME)-$(VERSION).gem
AMD64_DEB=pkg/$(NAME)_$(VERSION)_amd64.deb
ARM64_DEB=pkg/$(NAME)_$(VERSION)_arm64.deb

GOFILES=$(shell find . -type f -name '*.go')
MANFILES=$(shell find man -name '*.ronn' -exec echo build/{} \; | sed 's/\.ronn/\.gz/')

BUNDLE_EXEC=bundle exec
SHELL=/usr/bin/env bash

export GO111MODULE=on

.PHONY: default all binaries gem man clean dev_bootstrap setup

default: all
all: setup gem deb
binaries: \
	build/bin/linux-amd64 \
	build/bin/linux-arm64 \
	build/bin/darwin-amd64 \
	build/bin/darwin-arm64 \
	build/bin/freebsd-amd64 \
	build/bin/windows-amd64.exe
gem: $(GEM)
deb: $(AMD64_DEB) $(ARM64_DEB)
man: $(MANFILES)

build/man/%.gz: man/%.ronn
	mkdir -p "$(@D)"
	set -euo pipefail ; $(BUNDLE_EXEC) ronn -r --pipe "$<" | gzip > "$@" || (rm -f "$<" ; false)

build/bin/linux-amd64: $(GOFILES) cmd/$(NAME)/version.go
	GOOS=linux GOARCH=amd64 go build -o "$@" "$(PACKAGE)/cmd/$(NAME)"
build/bin/linux-arm64: $(GOFILES) cmd/$(NAME)/version.go
	GOOS=linux GOARCH=arm64 go build -o "$@" "$(PACKAGE)/cmd/$(NAME)"
build/bin/darwin-amd64: $(GOFILES) cmd/$(NAME)/version.go
	GOOS=darwin GOARCH=amd64 go build -o "$@" "$(PACKAGE)/cmd/$(NAME)"
build/bin/darwin-arm64: $(GOFILES) cmd/$(NAME)/version.go
	GOOS=darwin GOARCH=arm64 go build -o "$@" "$(PACKAGE)/cmd/$(NAME)"
build/bin/freebsd-amd64: $(GOFILES) cmd/$(NAME)/version.go
	GOOS=freebsd GOARCH=amd64 go build -o "$@" "$(PACKAGE)/cmd/$(NAME)"
build/bin/windows-amd64.exe: $(GOFILES) cmd/$(NAME)/version.go
	GOOS=windows GOARCH=amd64 go build -o "$@" "$(PACKAGE)/cmd/$(NAME)"

$(GEM): rubygem/$(NAME)-$(VERSION).gem
	mkdir -p $(@D)
	mv "$<" "$@"

rubygem/$(NAME)-$(VERSION).gem: \
	rubygem/lib/$(NAME)/version.rb \
	rubygem/build/linux-amd64/ejson \
	rubygem/build/linux-arm64/ejson \
	rubygem/LICENSE.txt \
	rubygem/build/darwin-amd64/ejson \
	rubygem/build/darwin-arm64/ejson \
	rubygem/build/freebsd-amd64/ejson \
	rubygem/build/windows-amd64/ejson.exe \
	rubygem/man
	cd rubygem && gem build ejson.gemspec

rubygem/LICENSE.txt: LICENSE.txt
	cp "$<" "$@"

rubygem/man: man
	cp -a build/man $@

rubygem/build/darwin-amd64/ejson: build/bin/darwin-amd64
	mkdir -p $(@D)
	cp -a "$<" "$@"

rubygem/build/darwin-arm64/ejson: build/bin/darwin-arm64
	mkdir -p $(@D)
	cp -a "$<" "$@"

rubygem/build/freebsd-amd64/ejson: build/bin/freebsd-amd64
	mkdir -p $(@D)
	cp -a "$<" "$@"

rubygem/build/linux-amd64/ejson: build/bin/linux-amd64
	mkdir -p $(@D)
	cp -a "$<" "$@"

rubygem/build/linux-arm64/ejson: build/bin/linux-arm64
	mkdir -p $(@D)
	cp -a "$<" "$@"

rubygem/build/windows-amd64/ejson.exe: build/bin/windows-amd64.exe
	mkdir -p $(@D)
	cp -a "$<" "$@"

cmd/$(NAME)/version.go: VERSION
	printf '%b' 'package main\n\nconst VERSION string = "$(VERSION)"\n' > $@

rubygem/lib/$(NAME)/version.rb: VERSION
	mkdir -p $(@D)
	printf '%b' 'module $(RUBY_MODULE)\n  VERSION = "$(VERSION)"\nend\n' > $@

$(AMD64_DEB): build/bin/linux-amd64 man
	mkdir -p $(@D)
	rm -f "$@"
	$(BUNDLE_EXEC) fpm \
		-t deb \
		-s dir \
		--name="$(NAME)" \
		--version="$(VERSION)" \
		--package="$@" \
		--license=MIT \
		--category=admin \
		--no-depends \
		--no-auto-depends \
		--architecture=amd64 \
		--maintainer="Shopify <admins@shopify.com>" \
		--description="utility for managing a collection of secrets in source control. Secrets are encrypted using public key, elliptic curve cryptography." \
		--url="https://github.com/Shopify/ejson" \
		./build/man/=/usr/share/man/ \
		./$<=/usr/bin/$(NAME)

$(ARM64_DEB): build/bin/linux-arm64 man
	mkdir -p $(@D)
	rm -f "$@"
	$(BUNDLE_EXEC) fpm \
		-t deb \
		-s dir \
		--name="$(NAME)" \
		--version="$(VERSION)" \
		--package="$@" \
		--license=MIT \
		--category=admin \
		--no-depends \
		--no-auto-depends \
		--architecture=arm64 \
		--maintainer="Shopify <admins@shopify.com>" \
		--description="utility for managing a collection of secrets in source control. Secrets are encrypted using public key, elliptic curve cryptography." \
		--url="https://github.com/Shopify/ejson" \
		./build/man/=/usr/share/man/ \
		./$<=/usr/bin/$(NAME)

setup:
	go mod download
	go mod tidy

clean:
	rm -rf build pkg rubygem/{LICENSE.txt,lib/ejson/version.rb,build,*.gem}
