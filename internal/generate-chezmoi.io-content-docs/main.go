package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
)

var (
	debug      = flag.Bool("debug", false, "debug")
	shortTitle = flag.String("shorttitle", "", "short title")
	longTitle  = flag.String("longtitle", "", "long title")

	replaceURLRegexps = map[string]*regexp.Regexp{
		"/docs/changes/":      regexp.MustCompile(`https://github.com/twpayne/chezmoi/blob/master/docs/CHANGES.md`),
		"/docs/contributing/": regexp.MustCompile(`https://github.com/twpayne/chezmoi/blob/master/docs/CONTRIBUTING.md`),
		"/docs/faq/":          regexp.MustCompile(`https://github.com/twpayne/chezmoi/blob/master/docs/FAQ.md`),
		"/docs/how-to/":       regexp.MustCompile(`https://github.com/twpayne/chezmoi/blob/master/docs/HOWTO.md`),
		"/docs/install/":      regexp.MustCompile(`https://github.com/twpayne/chezmoi/blob/master/docs/INSTALL.md`),
		"/docs/quick-start/":  regexp.MustCompile(`https://github.com/twpayne/chezmoi/blob/master/docs/QUICKSTART.md`),
		"/docs/reference/":    regexp.MustCompile(`https://github.com/twpayne/chezmoi/blob/master/docs/REFERENCE.md`),
	}
)

func run() error {
	flag.Parse()

	fmt.Printf(""+
		"---\n"+
		"title: %q\n"+
		"---\n"+
		"\n",
		*shortTitle,
	)

	s := bufio.NewScanner(os.Stdin)
	state := "replace-title"
	for s.Scan() {
		if *debug {
			log.Printf("%s: %q", state, s.Text())
		}
		switch state {
		case "replace-title":
			fmt.Printf("# %s\n\n", *longTitle)
			state = "find-toc"
		case "find-toc":
			if s.Text() == "<!--- toc --->" {
				state = "skip-toc"
			}
		case "skip-toc":
			if s.Text() == "" {
				state = "copy-content"
			}
		case "copy-content":
			text := s.Text()
			for docsURL, re := range replaceURLRegexps {
				text = re.ReplaceAllString(text, docsURL)
			}
			fmt.Println(text)
		}
	}
	return s.Err()
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
