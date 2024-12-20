package archivetest

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"slices"

	"github.com/klauspost/compress/zip"
)

func NewZip(root map[string]any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buffer)
	for _, key := range slices.Sorted(maps.Keys(root)) {
		if err := zipAddEntry(zipWriter, key, root[key]); err != nil {
			return nil, err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func zipAddEntry(w *zip.Writer, name string, entry any) error {
	switch entry := entry.(type) {
	case []byte:
		return zipAddEntryFile(w, name, entry, 0o666)
	case map[string]any:
		return zipAddEntryDir(w, name, fs.ModePerm, entry)
	case string:
		return zipAddEntryFile(w, name, []byte(entry), 0o666)
	case *Dir:
		return zipAddEntryDir(w, name, entry.Perm, entry.Entries)
	case *File:
		return zipAddEntryFile(w, name, entry.Contents, entry.Perm)
	case *Symlink:
		return zipAddEntrySymlink(w, name, []byte(entry.Target))
	default:
		return fmt.Errorf("%s: unsupported type: %T", name, entry)
	}
}

func zipAddEntryDir(w *zip.Writer, name string, perm fs.FileMode, entries map[string]any) error {
	fileHeader := zip.FileHeader{
		Name: name,
	}
	fileHeader.SetMode(fs.ModeDir | perm)
	if _, err := w.CreateHeader(&fileHeader); err != nil {
		return err
	}
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		if err := zipAddEntry(w, name+"/"+key, entries[key]); err != nil {
			return err
		}
	}
	return nil
}

func zipAddEntryFile(w *zip.Writer, name string, data []byte, perm fs.FileMode) error {
	fileHeader := zip.FileHeader{
		Name:               name,
		Method:             zip.Deflate,
		UncompressedSize64: uint64(len(data)),
	}
	fileHeader.SetMode(perm)
	fileWriter, err := w.CreateHeader(&fileHeader)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(data)
	return err
}

func zipAddEntrySymlink(w *zip.Writer, name string, target []byte) error {
	fileHeader := zip.FileHeader{
		Name:               name,
		UncompressedSize64: uint64(len(target)),
	}
	fileHeader.SetMode(fs.ModeSymlink)
	fileWriter, err := w.CreateHeader(&fileHeader)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(target)
	return err
}
