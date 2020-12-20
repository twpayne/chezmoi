.PHONY: website
website: content-docs
	./make-gh-pages.sh

.PHONY: content-docs
content-docs: \
	content/docs/changes.md \
	content/docs/contributing.md \
	content/docs/faq.md \
	content/docs/how-to.md \
	content/docs/install.md \
	content/docs/media.md \
	content/docs/quick-start.md \
	content/docs/reference.md \
	content/docs/templating.md

content/docs/changes.md: ../docs/CHANGES.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="Changes" -longtitle="Changes" < $< > $@ || ( rm -f $@ ; false )

content/docs/contributing.md: ../docs/CONTRIBUTING.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="Contributing" -longtitle="Contributing Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/faq.md: ../docs/FAQ.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="FAQ" -longtitle="Frequently Asked Questions" < $< > $@ || ( rm -f $@ ; false )

content/docs/how-to.md: ../docs/HOWTO.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="How-To" -longtitle="How-To Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/install.md: ../docs/INSTALL.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="Install" -longtitle="Install Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/media.md: ../docs/MEDIA.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="Media" -longtitle="Media" < $< > $@ || ( rm -f $@ ; false )

content/docs/quick-start.md: ../docs/QUICKSTART.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="Quick Start" -longtitle="Quick Start Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/reference.md: ../docs/REFERENCE.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="Reference" -longtitle="Reference Manual" < $< > $@ || ( rm -f $@ ; false )

content/docs/templating.md: ../docs/TEMPLATING.md ../internal/cmd/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../internal/cmd/generate-chezmoi.io-content-docs -shorttitle="Templating" -longtitle="Templating Guide" < $< > $@ || ( rm -f $@ ; false )
