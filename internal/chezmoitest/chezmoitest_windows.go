//go:build windows && test

package chezmoitest

// umaskStr is the umask used in tests represented as a string so it can be
// set with the
//
//	-ldflags="-X github.com/twpayne/chezmoi/v2/internal/chezmoitest.umaskStr=..."
//
// option to go build and go test.
var umaskStr = "0"

func init() {
	// Umask is the umask used in tests.
	//
	// On Windows, Umask is zero as Windows does not use POSIX-style permissions.
	Umask = mustParseFileMode(umaskStr)
}
