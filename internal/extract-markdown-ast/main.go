package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/russross/blackfriday/v2"
)

func isContainer(nt blackfriday.NodeType) bool {
	switch nt {
	case blackfriday.Document:
		fallthrough
	case blackfriday.BlockQuote:
		fallthrough
	case blackfriday.List:
		fallthrough
	case blackfriday.Item:
		fallthrough
	case blackfriday.Paragraph:
		fallthrough
	case blackfriday.Heading:
		fallthrough
	case blackfriday.Emph:
		fallthrough
	case blackfriday.Strong:
		fallthrough
	case blackfriday.Del:
		fallthrough
	case blackfriday.Link:
		fallthrough
	case blackfriday.Image:
		fallthrough
	case blackfriday.Table:
		fallthrough
	case blackfriday.TableHead:
		fallthrough
	case blackfriday.TableBody:
		fallthrough
	case blackfriday.TableRow:
		fallthrough
	case blackfriday.TableCell:
		return true
	default:
		return false
	}
}

func run() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	depth := 0
	b := blackfriday.New(blackfriday.WithExtensions(blackfriday.Tables))
	b.Parse(data).Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if isContainer(node.Type) {
			if entering {
				fmt.Printf("%s%s {\n", strings.Repeat("\t", depth), node.Type.String())
				depth++
			} else {
				depth--
				fmt.Printf("%s}\n", strings.Repeat("\t", depth))
			}
		} else {
			fmt.Printf("%s%s %q\n", strings.Repeat("\t", depth), node.Type.String(), node.Literal)
		}
		return blackfriday.GoToNext
	})
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
