package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs"
)

func getKernelInfo(fs vfs.FS) (map[string]string, error) {
	const procKernel = "/proc/sys/kernel"
	files := []string{"version", "ostype", "osrelease"}

	stat, err := fs.Stat(procKernel)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("expected %q to be a directory", procKernel)
	}

	res := map[string]string{}
	for _, file := range files {
		p := filepath.Join(procKernel, file)
		f, err := fs.Open(p)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		res[file] = strings.TrimSpace(string(data))
	}

	return res, nil
}
