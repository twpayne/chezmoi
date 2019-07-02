// +build windows

package cmd

import (
    "os"
)

// TODO: what should this really do?
func getOwner(info os.FileInfo) int {
    return -1
}
