// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package tests

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TempDir(t *testing.T) string {
	t.Helper()

	dir, err := ioutil.TempDir("", "logstore")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})

	return dir
}
