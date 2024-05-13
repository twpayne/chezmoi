package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	output := flag.String("o", "", "output file")
	flag.Parse()

	b := strings.Builder{}
	sha, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		log.Fatal(err)
	}
	b.Write(bytes.TrimSpace(sha))

	if err := exec.Command("git", "diff-index", "--quiet", "HEAD").Run(); err != nil {
		b.WriteString("-dirty")
	}

	commit := b.String()

	if *output != "" {
		err := os.WriteFile(*output, []byte(commit), 0o644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(commit)
	}
}
