package cmd

import (
    "runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSecretFunc(t *testing.T) {
	t.Parallel()
	c := &Config{
		GenericSecret: genericSecretCmdConfig{
			Command: "date",
		},
	}
	args := []string{"+%Y-%M-%DT%H:%M:%SZ"}
	var value interface{}
	assert.NotPanics(t, func() {
		value = c.secretFunc(args...)
	})
	time.Sleep(1100 * time.Millisecond)
	assert.Equal(t, value, c.secretFunc(args...))
}

func TestSecretJSONFunc(t *testing.T) {
	t.Parallel()

	var c *Config
	var args []string

	if runtime.GOOS != "windows" {
        c = &Config{
            GenericSecret: genericSecretCmdConfig{
                Command: "date",
            },
        }
        args = []string{`+{"date":"%Y-%M-%DT%H:%M:%SZ"}`}
    } else {
        // Windows doesn't (usually) have "date", but powershell is included with
        // all versions of Windows v7 or newer.
        c = &Config{
            GenericSecret: genericSecretCmdConfig{
                Command: "powershell.exe",
            },
        }
        args = []string{"-Command", "Get-Date | ConvertTo-Json"}
    }

	var value interface{}
	assert.NotPanics(t, func() {
		value = c.secretJSONFunc(args...)
	})
	time.Sleep(1100 * time.Millisecond)
	assert.Equal(t, value, c.secretJSONFunc(args...))
}
