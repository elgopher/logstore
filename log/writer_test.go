// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log_test

import (
	"testing"

	"github.com/jacekolszak/logstore/internal/tests"
	"github.com/jacekolszak/logstore/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_Write(t *testing.T) {
	t.Run("should generate entry time", func(t *testing.T) {
		now := time2006(t)
		writer := tests.OpenLogWriter(t, log.NowFunc(fixedNow(now)))
		// when
		entryTime, err := writer.Write(data1)
		// then
		require.NoError(t, err)
		assert.Equal(t, now, entryTime)
	})

	t.Run("should increase time artificially when time has not changed", func(t *testing.T) {
		now := time2006(t)
		writer := tests.OpenLogWriter(t, log.NowFunc(fixedNow(now)))
		// when
		t1, err1 := writer.Write(data1)
		t2, err2 := writer.Write(data2)
		// then
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.True(t, t1.Equal(now))
		assert.True(t, t2.After(now))
	})

	t.Run("should increase time artificially when time has not changed and writer was re-open", func(t *testing.T) {
		now := time2006(t)
		l, writer := tests.OpenLogWithWriter(t, log.NowFunc(fixedNow(now)))
		_, _ = writer.Write(data1)
		_ = writer.Close()
		writer, err := l.OpenWriter(log.NowFunc(fixedNow(now)))
		defer tests.CloseCloser(t, writer)
		require.NoError(t, err)
		// when
		t2, err := writer.Write(data2)
		// then
		require.NoError(t, err)
		assert.True(t, t2.After(now))
	})

	t.Run("should increase time artificially when time has gone back", func(t *testing.T) {
		t1 := time2006(t)
		t2 := time2005(t)
		clock := &clock{
			currentTime: &t1,
		}
		writer := tests.OpenLogWriter(t, log.NowFunc(clock.Now))
		// when
		actualTime1, err1 := writer.Write(data1)
		clock.currentTime = &t2
		actualTime2, err2 := writer.Write(data2)
		// then
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.True(t, actualTime2.After(actualTime1))
	})

	t.Run("should append entry to log", func(t *testing.T) {
		now := time2006(t)
		l, writer := tests.OpenLogWithWriter(t, log.NowFunc(fixedNow(now)))
		// when
		_, err := writer.Write(data1)
		// then
		require.NoError(t, err)
		entries := tests.ReadAll(t, l)
		assert.Equal(t,
			[]tests.Entry{
				{Time: now, Data: data1},
			},
			entries)
	})

	t.Run("should append 2 entries", func(t *testing.T) {
		now := time2006(t)
		clock := &clock{
			currentTime: &now,
		}
		l, writer := tests.OpenLogWithWriter(t, log.NowFunc(clock.Now))
		// when
		_, err1 := writer.Write(data1)
		clock.moveForward()
		t2, err2 := writer.Write(data2)
		// then
		require.NoError(t, err1)
		require.NoError(t, err2)
		entries := tests.ReadAll(t, l)
		assert.Equal(t,
			[]tests.Entry{
				{Time: now, Data: data1},
				{Time: t2, Data: data2},
			},
			entries)
	})
}
