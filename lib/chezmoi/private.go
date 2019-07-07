package chezmoi

import "os"

type PrivacyStater interface {
	RawPath(name string) (string, error)
	Stat(name string) (os.FileInfo, error)
}
