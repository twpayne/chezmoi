package xdg

import "os/exec"

// Open opens fileOrURL with xdg-open.
func Open(fileOrURL string) error {
	return exec.Command("xdg-open", fileOrURL).Run()
}
