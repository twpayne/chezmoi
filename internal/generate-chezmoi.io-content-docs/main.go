package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	debug      = flag.Bool("debug", false, "debug")
	shortTitle = flag.String("shorttitle", "", "short title")
	longTitle  = flag.String("longtitle", "", "long title")
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
			fmt.Println(s.Text())
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
