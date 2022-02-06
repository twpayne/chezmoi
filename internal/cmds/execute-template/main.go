package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/renameio/v2/maybe"
	"gopkg.in/yaml.v3"
)

var (
	templateDataFilename = flag.String("data", "", "data filename")
	outputFilename       = flag.String("output", "", "output filename")
)

func run() error {
	flag.Parse()

	var templateData interface{}
	if *templateDataFilename != "" {
		dataBytes, err := os.ReadFile(*templateDataFilename)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(dataBytes, &templateData); err != nil {
			return err
		}
	}

	if flag.NArg() == 0 {
		return fmt.Errorf("no arguments")
	}

	templateName := path.Base(flag.Arg(0))
	buffer := &bytes.Buffer{}
	tmpl, err := template.New(templateName).Funcs(sprig.TxtFuncMap()).ParseFiles(flag.Args()...)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(buffer, templateData); err != nil {
		return err
	}

	if *outputFilename == "" {
		if _, err := os.Stdout.Write(buffer.Bytes()); err != nil {
			return err
		}
	} else if data, err := os.ReadFile(*outputFilename); err != nil || !bytes.Equal(data, buffer.Bytes()) {
		if err := maybe.WriteFile(*outputFilename, buffer.Bytes(), 0o666); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
