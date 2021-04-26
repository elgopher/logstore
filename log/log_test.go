// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log_test

import (
	"path"
	"testing"

	"github.com/jacekolszak/logstore/internal/tests"
	"github.com/jacekolszak/logstore/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	t.Run("should create directory", func(t *testing.T) {
		tmpDir := tests.TempDir(t)
		dir := path.Join(tmpDir, "missing")
		// when
		l, err := log.Open(dir)
		defer closeLog(t, l)
		// then
		require.NoError(t, err)
		assert.DirExists(t, dir)
	})

	t.Run("should open log", func(t *testing.T) {
		dir := tests.TempDir(t)
		// when
		l, err := log.Open(dir)
		defer closeLog(t, l)
		// then
		require.NoError(t, err)
		assert.NotNil(t, l)
	})

	t.Run("should return error for option returning error", func(t *testing.T) {
		dir := tests.TempDir(t)
		failingOption := func(l *log.Log) error {
			return stringError("error")
		}
		// when
		l, err := log.Open(dir, failingOption)
		defer closeLog(t, l)
		// then
		assert.Error(t, err)
		assert.Nil(t, l)
	})

	t.Run("should skip nil option", func(t *testing.T) {
		dir := tests.TempDir(t)
		// when
		l, err := log.Open(dir, nil)
		defer closeLog(t, l)
		// then
		require.NoError(t, err)
		assert.NotNil(t, l)
	})
}

func closeLog(t *testing.T, l *log.Log) {
	t.Helper()
	require.NoError(t, l.Close())
}

type stringError string

func (s stringError) Error() string {
	return string(s)
}
