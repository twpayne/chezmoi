package archive

import (
	"archive/tar"
	"bytes"
	"fmt"
	"sort"
)

// NewTar returns the bytes of a new tar archive containing root.
func NewTar(root map[string]interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(buffer)
	keys := make([]string, 0, len(root))
	for key := range root {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := tarAddEntry(tarWriter, key, root[key]); err != nil {
			return nil, err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func tarAddEntry(w *tar.Writer, name string, entry interface{}) error {
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

	case map[string]interface{}:
		if err := w.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     name + "/",
			Mode:     0o777,
		}); err != nil {
			return err
		}
		for key, value := range entry {
			if err := tarAddEntry(w, name+"/"+key, value); err != nil {
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
		for key, value := range entry.Entries {
			if err := tarAddEntry(w, name+"/"+key, value); err != nil {
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
