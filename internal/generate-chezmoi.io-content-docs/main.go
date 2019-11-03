package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	shortTitle = flag.String("shorttitle", "", "short title")
	longTitle  = flag.String("longtitle", "", "long title")
)

func run() error {
	flag.Parse()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	fmt.Printf(
		"+++\n"+
			"title = %q\n"+
			"+++\n"+
			"\n"+
			"# %s\n",
		*shortTitle,
		*longTitle,
	)

	if index := bytes.IndexByte(data, '\n'); index != -1 {
		os.Stdout.Write(data[index+1:])
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
