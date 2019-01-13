package cmd

import "strconv"

// An octalIntValue is an int that is printed in octal. It implements the
// pflag.Value interface for use as a command line flag.
type octalIntValue int

func (o *octalIntValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*o = octalIntValue(v)
	return err
}

func (o *octalIntValue) String() string {
	return "0" + strconv.FormatInt(int64(*o), 8)
}

func (o *octalIntValue) Type() string {
	return "int"
}
