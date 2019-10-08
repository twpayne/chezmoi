package chezmoi

import "os"

// A PrivacyStater implements the minimum methods required for IsPrivate.
type PrivacyStater interface {
	RawPath(name string) (string, error)
	Stat(name string) (os.FileInfo, error)
}
