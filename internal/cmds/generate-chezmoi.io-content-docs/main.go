package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	debug      = flag.Bool("debug", false, "debug")
	shortTitle = flag.String("shorttitle", "", "short title")
	longTitle  = flag.String("longtitle", "", "long title")

	replaceURLRegexp       = regexp.MustCompile(`https://github\.com/twpayne/chezmoi/blob/master/docs/[A-Z]+\.md`) // lgtm[go/regex/missing-regexp-anchor]
	nonStandardPageRenames = map[string]string{
		"HOWTO":      "how-to",
		"QUICKSTART": "quick-start",
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
			text = replaceURLRegexp.ReplaceAllStringFunc(text, func(s string) string {
				name := path.Base(s)
				name = strings.TrimSuffix(name, path.Ext(name))
				var ok bool
				newName, ok := nonStandardPageRenames[name]
				if !ok {
					newName = strings.ToLower(name)
				}
				return "/docs/" + newName + "/"
			})
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
