// +build snap

package cmd

import "os"

// snapFix, when built with snap, applies any fixes required to work with snap.
func (c *Config) snapFix() error {
	// Snap sets the XDG_RUNTIME_DIR environment variable to
	// /run/user/$uid/snap.$snap_name, but does create this directory.
	// Consequently, any spawned processes that need $XDG_DATA_DIR will fail. As
	// a work-around, create the directory if it does not exist. See
	// https://forum.snapcraft.io/t/wayland-dconf-and-xdg-runtime-dir/186/13.
	return os.MkdirAll(c.bds.RuntimeDir, 0o700&^os.FileMode(c.Umask))
}
