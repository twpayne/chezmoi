//+build !windows

package cmd

import "syscall"

func getUmask() int {
	// FIXME should we call runtime.LockOSThread or similar?
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	return umask
}
