// +build windows

package chezmoi

import (
	"archive/tar"
	"os/user"
	"time"
)

func (ts *TargetState) getTarHeaderTemplate() (*tar.Header, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &tar.Header{
		Uname:      currentUser.Username,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}, nil
}
