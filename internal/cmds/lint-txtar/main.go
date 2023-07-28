package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rogpeppe/go-internal/txtar"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
)

var write = flag.Bool("w", false, "rewrite archives")

func lintFilenames(archiveFilename string, archive *txtar.Archive) error {
	var errs error
	filenames := make(map[string]struct{})
	for _, file := range archive.Files {
		if file.Name == "" {
			errs = multierr.Append(errs, fmt.Errorf("%s: empty filename", archiveFilename))
		} else {
			if _, ok := filenames[file.Name]; ok {
				errs = multierr.Append(errs, fmt.Errorf("%s: %s: duplicate filename", archiveFilename, file.Name))
			}
			filenames[file.Name] = struct{}{}
		}
	}
	return errs
}

func sortFilesFunc(file1, file2 txtar.File) int {
	fileComponents1 := strings.Split(file1.Name, "/")
	fileComponents2 := strings.Split(file2.Name, "/")
	return slices.Compare(fileComponents1, fileComponents2)
}

func tidyTxtar(archiveFilename string) error {
	archive, err := txtar.ParseFile(archiveFilename)
	if err != nil {
		return err
	}

	if err := lintFilenames(archiveFilename, archive); err != nil {
		return err
	}

	if slices.IsSortedFunc(archive.Files, sortFilesFunc) {
		return nil
	}

	if *write {
		slices.SortFunc(archive.Files, sortFilesFunc)
		return os.WriteFile(archiveFilename, txtar.Format(archive), 0o666)
	}

	return fmt.Errorf("%s: files are not sorted", archiveFilename)
}

func run() error {
	flag.Parse()

	var errs error
	for _, arg := range flag.Args() {
		errs = multierr.Append(errs, tidyTxtar(arg))
	}
	return errs
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
