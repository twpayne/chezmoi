.PHONY: website
website: content-docs
	./make-gh-pages.sh

.PHONY: serve
serve:
	hugo serve

.PHONY: content-docs
content-docs: \
	content/docs/architecture.md \
	content/docs/changes.md \
	content/docs/comparison.md \
	content/docs/contributing.md \
	content/docs/faq.md \
	content/docs/how-to.md \
	content/docs/install.md \
	content/docs/media.md \
	content/docs/quick-start.md \
	content/docs/reference.md \
	content/docs/related.md \
	content/docs/security.md \
	content/docs/templating.md

content/docs/architecture.md: ../../docs/ARCHITECTURE.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Architecture" -longtitle="Architecture" < $< > $@ || ( rm -f $@ ; false )

content/docs/changes.md: ../../docs/CHANGES.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Changes" -longtitle="Changes" < $< > $@ || ( rm -f $@ ; false )

content/docs/comparison.md: ../../docs/COMPARISON.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Comparison" -longtitle="Comparison Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/contributing.md: ../../docs/CONTRIBUTING.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Contributing" -longtitle="Contributing Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/faq.md: ../../docs/FAQ.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="FAQ" -longtitle="Frequently Asked Questions" < $< > $@ || ( rm -f $@ ; false )

content/docs/how-to.md: ../../docs/HOWTO.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="How-To" -longtitle="How-To Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/install.md: ../../docs/INSTALL.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Install" -longtitle="Install Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/media.md: ../../docs/MEDIA.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Media" -longtitle="Media" < $< > $@ || ( rm -f $@ ; false )

content/docs/quick-start.md: ../../docs/QUICKSTART.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Quick Start" -longtitle="Quick Start Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/reference.md: ../../docs/REFERENCE.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Reference" -longtitle="Reference Manual" < $< > $@ || ( rm -f $@ ; false )

content/docs/related.md: ../../docs/RELATED.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Related" -longtitle="Related Software" < $< > $@ || ( rm -f $@ ; false )

content/docs/security.md: ../../docs/SECURITY.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Security" -longtitle="Security" < $< > $@ || ( rm -f $@ ; false )

content/docs/templating.md: ../../docs/TEMPLATING.md ../../internal/cmds/generate-chezmoi.io-content-docs/main.go Makefile
	go run ../../internal/cmds/generate-chezmoi.io-content-docs -shorttitle="Templating" -longtitle="Templating Guide" < $< > $@ || ( rm -f $@ ; false )
