//go:build unix && test

package chezmoitest

import "golang.org/x/sys/unix"

// umaskStr is the umask used in tests represented as a string so it can be
// set with the
//
//	-ldflags="-X github.com/twpayne/chezmoi/v2/internal/chezmoitest.umaskStr=..."
//
// option to go build and go test.
var umaskStr = "0o022"

func init() {
	// Umask is the umask used in tests.
	//
	// If you change this then you will need to update the testscripts in
	// testdata/scripts where permissions after applying umask are hardcoded as
	// strings. Pure Go tests should use this value to ensure that they pass,
	// irrespective of what it is set to. Be aware that the process's umask is a
	// process-level property and cannot be locally changed within individual
	// tests.
	Umask = mustParseFileMode(umaskStr)
	unix.Umask(int(Umask))
}
