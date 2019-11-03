.PHONY: content-docs
content-docs: \
	content/docs/changes.md \
	content/docs/contributing.md \
	content/docs/faq.md \
	content/docs/how-to.md \
	content/docs/install.md \
	content/docs/quick-start.md \
	content/docs/reference.md

content/docs/changes.md: ../docs/CHANGES.md Makefile
	go run ../internal/generate-chezmoi.io-content-docs -shorttitle="Changes" -longtitle="Changes" < $< > $@ || ( rm -f $@ ; false )

content/docs/contributing.md: ../docs/CONTRIBUTING.md Makefile
	go run ../internal/generate-chezmoi.io-content-docs -shorttitle="Contributing" -longtitle="Contributing Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/faq.md: ../docs/FAQ.md Makefile
	go run ../internal/generate-chezmoi.io-content-docs -shorttitle="FAQ" -longtitle="Frequently Asked Questions" < $< > $@ || ( rm -f $@ ; false )

content/docs/how-to.md: ../docs/HOWTO.md Makefile
	go run ../internal/generate-chezmoi.io-content-docs -shorttitle="How-To" -longtitle="How-To Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/install.md: ../docs/INSTALL.md Makefile
	go run ../internal/generate-chezmoi.io-content-docs -shorttitle="Install" -longtitle="Install Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/quick-start.md: ../docs/QUICKSTART.md Makefile
	go run ../internal/generate-chezmoi.io-content-docs -shorttitle="Quick Start" -longtitle="Quick Start Guide" < $< > $@ || ( rm -f $@ ; false )

content/docs/reference.md: ../docs/REFERENCE.md Makefile
	go run ../internal/generate-chezmoi.io-content-docs -shorttitle="Reference" -longtitle="Reference Manual" < $< > $@ || ( rm -f $@ ; false )

