package archivetest

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/fs"
)

// NewTar returns the bytes of a new tar archive containing root.
func NewTar(root map[string]any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(buffer)
	for _, key := range sortedKeys(root) {
		if err := tarAddEntry(tarWriter, key, root[key]); err != nil {
			return nil, err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func tarAddEntry(w *tar.Writer, name string, entry any) error {
	switch entry := entry.(type) {
	case []byte:
		if err := w.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Size:     int64(len(entry)),
			Mode:     0o666,
		}); err != nil {
			return err
		}
		if _, err := w.Write(entry); err != nil {
			return err
		}

	case map[string]any:
		if err := w.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     name + "/",
			Mode:     int64(fs.ModePerm),
		}); err != nil {
			return err
		}
		for _, key := range sortedKeys(entry) {
			if err := tarAddEntry(w, name+"/"+key, entry[key]); err != nil {
				return err
			}
		}
	case nil:
		return nil
	case string:
		if err := w.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Size:     int64(len(entry)),
			Mode:     0o666,
		}); err != nil {
			return err
		}
		if _, err := w.Write([]byte(entry)); err != nil {
			return err
		}
	case *Dir:
		if err := w.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     name + "/",
			Mode:     int64(entry.Perm),
		}); err != nil {
			return err
		}
		for _, key := range sortedKeys(entry.Entries) {
			if err := tarAddEntry(w, name+"/"+key, entry.Entries[key]); err != nil {
				return err
			}
		}
	case *File:
		if err := w.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Size:     int64(len(entry.Contents)),
			Mode:     int64(entry.Perm),
		}); err != nil {
			return err
		}
		if _, err := w.Write(entry.Contents); err != nil {
			return err
		}
	case *Symlink:
		if err := w.WriteHeader(&tar.Header{
			Typeflag: tar.TypeSymlink,
			Name:     name,
			Linkname: entry.Target,
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unsupported type: %T", name, entry)
	}
	return nil
}
