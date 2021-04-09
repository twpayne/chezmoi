package chezmoitest

var (
	umaskStr = "0"

	// Umask is the umask used in tests.
	Umask = mustParseFilemode(umaskStr)
)
