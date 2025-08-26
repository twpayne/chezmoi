// hexencode encodes stdin as hex. It is designed for use with the hexdecode
// testscript function.
package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func run() error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	for i := 0; i < len(data); i += 40 {
		line := data[i:min(i+40, len(data))]
		if _, err := fmt.Println(hex.EncodeToString(line)); err != nil {
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
