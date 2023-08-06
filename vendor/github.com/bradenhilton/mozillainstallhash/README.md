# mozillainstallhash

Gets the hash used to differentiate between installs of Mozilla software in `installs.ini` and `profiles.ini`.

### Example

1. Create a new local Go module:
   1. Create a new directory e.g. `$GOPATH/src/get_mozilla_install_hash`
   2. Change into the newly created directory
   3. Run `go mod init get_mozilla_install_hash`
2. Create a new file `get_mozilla_install_hash.go`:

	```go
	package main

	import (
		"fmt"
		"log"
		"os"
		"strings"

		"github.com/bradenhilton/mozillainstallhash"
	)

	const usage = `
	get_mozilla_install_hash
		Get the hash used to differentiate between installs of Mozilla software.

	Usage:
		get_mozilla_install_hash <path> [<path> ...]

	Where <path> is a string describing the parent directory of the executable,
	e.g. "C:\Program Files\Mozilla Firefox", with platform specific path separators
	("\" on Windows and "/" on Unix-like systems)

	Example:
		get_mozilla_install_hash "C:\Program Files\Mozilla Firefox"
		308046B0AF4A39CB

		get_mozilla_install_hash "C:/Program Files/Mozilla Firefox"
		9D561FCD08DC6D55

		get_mozilla_install_hash "/usr/lib/firefox"
		4F96D1932A9F858E`

	func main() {
		if len(os.Args) == 1 {
			log.Println(fmt.Errorf("error: no path specified"))
			fmt.Println(usage)
			os.Exit(1)
		}

		paths := os.Args[1:]
		for _, path := range paths {
			path = strings.TrimSuffix(path, "/")
			path = strings.TrimSuffix(path, "\\")

			hash, err := mozillainstallhash.MozillaInstallHash(path)
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Println(hash)
		}
	}
	```

3. Run `go get github.com/bradenhilton/mozillainstallhash`
4. Run `go run get_mozilla_install_hash.go <your_path_here>`

### License

MIT