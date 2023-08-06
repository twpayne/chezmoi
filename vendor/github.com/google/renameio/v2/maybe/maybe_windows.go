package maybe

import (
	"io/ioutil"
	"os"
)

// WriteFile mirrors ioutil.WriteFile. On Linux it uses renameio.WriteFile to
// create or replace an existing file with the same name atomically. On Windows
// files cannot be written atomically, so this function falls back to
// ioutil.WriteFile, which does not write the file atomically and ignores most
// permission bits. See https://github.com/google/renameio/issues/1 and
// https://github.com/golang/go/issues/22397#issuecomment-498856679 for
// discussion.
//
// Prefer using renameio.WriteFile instead so that you get an error if atomic
// replacement is not possible on the runtime platform. maybe.WriteFile is meant
// as a convenience wrapper if you are okay with atomic replacement not being
// supported by the runtime platform.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}
