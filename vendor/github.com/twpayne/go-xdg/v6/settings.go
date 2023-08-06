package xdg

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	settingsCmdName = "xdg-settings"

	// DefaultURLSchemeHandlerProperty is the default URL scheme handler property.
	DefaultURLSchemeHandlerProperty = "default-url-scheme-handler"

	// DefaultWebBrowserProperty is the default web browser property.
	DefaultWebBrowserProperty = "default-web-browser"
)

// A Setting is a setting.
type Setting struct {
	Property    string
	SubProperty string
}

// Check checks that value of s is value. See
// https://portland.freedesktop.org/doc/xdg-settings.html.
func (s Setting) Check(value string) (bool, error) {
	args := []string{"check", s.Property}
	if s.SubProperty != "" {
		args = append(args, s.SubProperty)
	}
	args = append(args, value)
	output, err := exec.Command(settingsCmdName, args...).Output()
	if err != nil {
		return false, err
	}
	o := strings.TrimSpace(string(output))
	switch o {
	case "yes":
		return true, nil
	case "no":
		return false, nil
	default:
		return false, fmt.Errorf(`%s %s: expected "yes" or "no", got %q`, settingsCmdName, strings.Join(args, " "), o)
	}
}

// Get gets the value of s.
func (s Setting) Get() (string, error) {
	args := []string{"get", s.Property}
	if s.SubProperty != "" {
		args = append(args, s.SubProperty)
	}
	output, err := exec.Command(settingsCmdName, args...).Output()
	return strings.TrimSpace(string(output)), err
}

// Set sets s to value.
func (s Setting) Set(value string) error {
	args := []string{"set", s.Property}
	if s.SubProperty != "" {
		args = append(args, s.SubProperty)
	}
	args = append(args, value)
	return exec.Command(settingsCmdName, args...).Run()
}
