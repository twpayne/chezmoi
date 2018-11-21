package absfstesting

import (
	"os"
	"path/filepath"

	"github.com/absfs/afero"
)

func MakeMemMapFs(fsMap map[string]string) (*afero.MemMapFs, error) {
	//fs := afero.NewMemMapFs()
	fs := &afero.MemMapFs{}
	for path, contents := range fsMap {
		if err := fs.MkdirAll(filepath.Dir(path), os.FileMode(0777)); err != nil {
			return nil, err
		}
		if err := afero.WriteFile(fs, path, []byte(contents), os.FileMode(0666)); err != nil {
			return nil, err
		}
	}
	return fs, nil
}

func MakeMapFs(fs afero.Fs) (map[string]string, error) {
	mapFs := make(map[string]string)
	if err := afero.Walk(fs, "/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		contents, err := afero.ReadFile(fs, path)
		if err != nil {
			return err
		}
		mapFs[path] = string(contents)
		return nil
	}); err != nil {
		return nil, err
	}
	return mapFs, nil
}
