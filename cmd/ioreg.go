//+build !darwin

package cmd

//nolint:unused
type ioregData struct{}

func init() {
	config.addTemplateFunc("ioreg", func() map[string]interface{} {
		return nil
	})
}
