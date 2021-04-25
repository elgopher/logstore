package tests

import (
	"testing"

	"github.com/jacekolszak/logstore/log"
	"github.com/stretchr/testify/require"
)

func OpenLog(t *testing.T, options ...log.OpenOption) *log.Log {
	s, err := log.Open(TempDir(t), options...)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = s.Close()
	})
	return s
}
