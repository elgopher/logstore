// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log_test

import (
	"testing"
	"time"

	"github.com/elgopher/logstore/internal/tests"
	"github.com/elgopher/logstore/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_WriteWithTime(t *testing.T) {
	t.Run("should append entry to log", func(t *testing.T) {
		now := time2006
		l, writer := tests.OpenLogWithWriter(t)
		// when
		err := writer.WriteWithTime(now, data1)
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
		l, writer := tests.OpenLogWithWriter(t)
		// when
		err1 := writer.WriteWithTime(time2005, data1)
		err2 := writer.WriteWithTime(time2006, data2)
		// then
		require.NoError(t, err1)
		require.NoError(t, err2)
		entries := tests.ReadAll(t, l)
		assert.Equal(t,
			[]tests.Entry{
				{Time: time2005, Data: data1},
				{Time: time2006, Data: data2},
			},
			entries)
	})

	t.Run("should create a new segment when segment max size is reached", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t, log.MaxSegmentSizeMB(1))
		oneMegabyteEntry := make([]byte, tests.OneMegabyte)
		_, err := writer.Write(oneMegabyteEntry)
		require.NoError(t, err)
		// when
		_, err = writer.Write(data1)
		// then
		require.NoError(t, err)
		segments, err := l.Segments()
		require.NoError(t, err)
		require.Len(t, segments, 2)
	})

	t.Run("segment file should be smaller than segment max size + at most size of last written entry", func(t *testing.T) {
		const (
			maxSegmentSizeMB         = 1
			entrySize          int64 = 1024
			segmentFileMaxSize       = int64(maxSegmentSizeMB)*tests.OneMegabyte + entrySize
		)

		t.Run("single writer", func(t *testing.T) {
			dir := tests.TempDir(t)
			l := log.New(dir)
			writer, err := l.OpenWriter(log.MaxSegmentSizeMB(maxSegmentSizeMB))
			defer tests.Close(t, writer)
			require.NoError(t, err)

			for i := 0; i < 1024*2; i++ {
				tests.WriteEntry(t, writer, entrySize)
			}
			// then
			tests.AssertFilesNoLargerThan(t, dir, segmentFileMaxSize)
		})

		t.Run("two writers", func(t *testing.T) {
			dir := tests.TempDir(t)
			l := log.New(dir)
			for j := 0; j < 2; j++ {
				writer, err := l.OpenWriter(log.MaxSegmentSizeMB(maxSegmentSizeMB))
				require.NoError(t, err)

				for i := 0; i < 1024; i++ {
					tests.WriteEntry(t, writer, entrySize)
				}

				tests.Close(t, writer)
			}
			// then
			tests.AssertFilesNoLargerThan(t, dir, segmentFileMaxSize)
		})
	})

	t.Run("should append entries to 2 segments", func(t *testing.T) {
		t1 := time2005
		t2 := time2005.Add(time.Hour)
		entry1 := make([]byte, tests.OneMegabyte)
		entry1[0] = 1
		entry2 := make([]byte, tests.OneMegabyte)
		entry2[0] = 2
		l, writer := tests.OpenLogWithWriter(t, log.MaxSegmentSizeMB(1))
		// when
		err1 := writer.WriteWithTime(t1, entry1)
		err2 := writer.WriteWithTime(t2, entry2)
		// then
		require.NoError(t, err1)
		require.NoError(t, err2)
		entries := tests.ReadAll(t, l)
		require.Len(t, entries, 2)
		assert.Equal(t, t1, entries[0].Time)
		assert.Equal(t, entry1, entries[0].Data)
		assert.Equal(t, t2, entries[1].Time)
		assert.Equal(t, entry2, entries[1].Data)
	})

	t.Run("should create new segment when max duration is reached", func(t *testing.T) {
		t1 := time2005
		t2 := t1.Add(time.Second).Add(time.Nanosecond)
		l, writer := tests.OpenLogWithWriter(t, log.MaxSegmentDuration(time.Second))
		// when
		_ = writer.WriteWithTime(t1, data1)
		_ = writer.WriteWithTime(t2, data2)
		// then
		segments, err := l.Segments()
		require.NoError(t, err)
		assert.Len(t, segments, 2)
	})

	t.Run("should return error when trying to force invalid time", func(t *testing.T) {
		cases := map[string]func(lastEntryTime time.Time) time.Time{
			"time before last entry": func(lastEntryTime time.Time) time.Time {
				return lastEntryTime.Add(-time.Nanosecond)
			},
			"time is equal last entry": func(lastEntryTime time.Time) time.Time {
				return lastEntryTime
			},
		}
		for name, givenTimeFunc := range cases {
			t.Run(name, func(t *testing.T) {
				l, writer := tests.OpenLogWithWriter(t)
				t1 := time2005
				_ = writer.WriteWithTime(t1, data1)
				givenTime := givenTimeFunc(t1)
				// when
				err := writer.WriteWithTime(givenTime, data2)
				// then
				assert.ErrorIs(t, err, log.ErrInvalidParameter)
				// and
				all := tests.ReadAll(t, l)
				assert.Len(t, all, 1)
			})
		}
	})
}

func TestWriter_Write(t *testing.T) {
	t.Run("should generate entry time", func(t *testing.T) {
		now := time2006
		writer := tests.OpenLogWriter(t, log.NowFunc(fixedNow(now)))
		// when
		entryTime, err := writer.Write(data1)
		// then
		require.NoError(t, err)
		assert.Equal(t, now, entryTime)
	})

	t.Run("should increase time artificially when time has not changed", func(t *testing.T) {
		now := time2006
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
		now := time2006
		l, writer := tests.OpenLogWithWriter(t, log.NowFunc(fixedNow(now)))
		_, _ = writer.Write(data1)
		_ = writer.Close()
		writer, err := l.OpenWriter(log.NowFunc(fixedNow(now)))
		defer tests.Close(t, writer)
		require.NoError(t, err)
		// when
		t2, err := writer.Write(data2)
		// then
		require.NoError(t, err)
		assert.True(t, t2.After(now))
	})

	t.Run("should increase time artificially when time has gone back", func(t *testing.T) {
		t1 := time2006
		t2 := time2005
		clock := &tests.Clock{
			CurrentTime: &t1,
		}
		writer := tests.OpenLogWriter(t, log.NowFunc(clock.Now))
		// when
		actualTime1, err1 := writer.Write(data1)
		clock.CurrentTime = &t2
		actualTime2, err2 := writer.Write(data2)
		// then
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.True(t, actualTime2.After(actualTime1))
	})
}
