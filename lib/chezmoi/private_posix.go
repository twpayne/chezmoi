// +build !windows

package chezmoi

// IsPrivate returns whether path should be considered private.
// nolint:interfacer
func IsPrivate(fs PrivacyStater, path string) (bool, error) {
	info, err := fs.Stat(path)
	if err != nil {
		return false, err
	}
	return info.Mode().Perm()&077 == 0, nil
}
