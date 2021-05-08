// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package tests

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TempDir(t TestingT) string {
	t.Helper()

	dir, err := ioutil.TempDir("", "logstore")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})

	return dir
}

func Close(t *testing.T, c io.Closer) {
	t.Helper()

	if c == nil || reflect.ValueOf(c).IsNil() {
		return
	}

	require.NoError(t, c.Close())
}
