// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertFilesNoLargerThan(t *testing.T, dir string, fileMaxSize int64) {
	t.Helper()

	dirEntries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			info, err := dirEntry.Info()
			require.NoError(t, err)

			assert.Truef(t, info.Size() < fileMaxSize,
				"%s file is too big (%d compared to max %d)", info.Name(), info.Size(), fileMaxSize)
		}
	}
}
