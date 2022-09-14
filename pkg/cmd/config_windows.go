package cmd

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func (c *Config) windowsVersion() (map[string]any, error) {
	registryKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("registry.OpenKey: %w", err)
	}
	windowsVersion := make(map[string]any)
	for _, name := range []string{
		"CurrentBuild",
		"CurrentVersion",
		"DisplayVersion",
		"EditionID",
		"ProductName",
	} {
		if value, _, err := registryKey.GetStringValue(name); err == nil {
			key := strings.ToLower(name[:1]) + name[1:]
			windowsVersion[key] = value
		}
	}
	for _, name := range []string{
		"CurrentMajorVersionNumber",
		"CurrentMinorVersionNumber",
	} {
		if value, _, err := registryKey.GetIntegerValue(name); err == nil {
			key := strings.ToLower(name[:1]) + name[1:]
			windowsVersion[key] = value
		}
	}
	return windowsVersion, nil
}
