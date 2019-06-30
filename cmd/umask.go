// +build !windows

package cmd

import "syscall"

func getUmask() int {
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	return umask
}
