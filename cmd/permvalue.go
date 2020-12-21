package cmd

import (
	"fmt"
	"strconv"
)

// An permValue is an int that is scanned and printed in octal. It implements
// the pflag.Value interface for use as a command line flag.
type permValue int

func (p *permValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 8, 64)
	if err != nil {
		return err
	}
	if v < 0o000 || 0o777 < v {
		return fmt.Errorf("%s: invalid permission value", s)
	}
	*p = permValue(v)
	return err
}

func (p *permValue) String() string {
	return fmt.Sprintf("%03o", *p)
}

func (p *permValue) Type() string {
	return "int"
}
