package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestGetKernelInfo(t *testing.T) {
	root := map[string]interface{}{
		"/proc/sys/kernel/version":   "#1 SMP Fri Nov 1 14:28:19 UTC 2019",
		"/proc/sys/kernel/ostype":    "Linux",
		"/proc/sys/kernel/osrelease": "4.19.81-microsoft-standard",
	}
	fs, cleanup, err := vfst.NewTestFS(root)
	require.NoError(t, err)
	defer cleanup()
	info, err := getKernelInfo(fs)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"version":   "#1 SMP Fri Nov 1 14:28:19 UTC 2019",
		"ostype":    "Linux",
		"osrelease": "4.19.81-microsoft-standard",
	}, info)
}
