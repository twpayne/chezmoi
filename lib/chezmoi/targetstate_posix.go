// +build !windows

package chezmoi

import (
	"archive/tar"
	"os/user"
	"strconv"
	"time"
)

func (ts *TargetState) getTarHeaderTemplate() (*tar.Header, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	uid, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.Atoi(currentUser.Gid)
	if err != nil {
		return nil, err
	}
	group, err := user.LookupGroupId(currentUser.Gid)
	if err != nil {
		return nil, err
	}

	return &tar.Header{
		Uid:        uid,
		Gid:        gid,
		Uname:      currentUser.Username,
		Gname:      group.Name,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}, nil
}
