//go:build !cgo

package shell

func cgoGetUserShell(name string) (string, bool) {
	return "", false
}
