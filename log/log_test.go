// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log_test

import (
	"path"
	"testing"
	"time"

	"github.com/elgopher/logstore/internal/tests"
	"github.com/elgopher/logstore/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	data1 = []byte("data1")
	data2 = []byte("data2")

	time2005 = tests.MustTime("2005-02-04T20:01:37Z")
	time2006 = tests.MustTime("2006-01-02T15:04:05Z")
)

func TestNew(t *testing.T) {
	t.Run("should create Log instance", func(t *testing.T) {
		l := log.New("dir")
		assert.NotNil(t, l)
	})
}

func TestLog_OpenReader(t *testing.T) {
	t.Run("should return error for option returning error", func(t *testing.T) {
		dir := tests.TempDir(t)
		failingOption := func(*log.ReaderSettings) error {
			return tests.ErrFixed
		}
		// when
		reader, err := log.New(dir).OpenReader(failingOption)
		defer tests.Close(t, reader)
		// then
		assert.ErrorIs(t, err, tests.ErrFixed)
		assert.Nil(t, reader)
	})

	t.Run("should skip nil option", func(t *testing.T) {
		dir := tests.TempDir(t)
		// when
		reader, err := log.New(dir).OpenReader(nil)
		defer tests.Close(t, reader)
		// then
		require.NoError(t, err)
		assert.NotNil(t, reader)
	})
}

func TestLog_OpenWriter(t *testing.T) {
	t.Run("should create directory", func(t *testing.T) {
		tmpDir := tests.TempDir(t)
		dir := path.Join(tmpDir, "missing")
		// when
		writer, err := log.New(dir).OpenWriter()
		defer tests.Close(t, writer)
		// then
		require.NoError(t, err)
		assert.DirExists(t, dir)
	})

	t.Run("should return error for option returning error", func(t *testing.T) {
		dir := tests.TempDir(t)
		failingOption := func(*log.WriterSettings) error {
			return tests.ErrFixed
		}
		// when
		writer, err := log.New(dir).OpenWriter(failingOption)
		defer tests.Close(t, writer)
		// then
		assert.ErrorIs(t, err, tests.ErrFixed)
		assert.Nil(t, writer)
	})

	t.Run("should skip nil option", func(t *testing.T) {
		dir := tests.TempDir(t)
		// when
		writer, err := log.New(dir).OpenWriter(nil)
		defer tests.Close(t, writer)
		// then
		require.NoError(t, err)
		assert.NotNil(t, writer)
	})

	t.Run("should return error when trying to open 2 writers simultaneously", func(t *testing.T) {
		dir := tests.TempDir(t)
		writer1, _ := log.New(dir).OpenWriter()
		defer tests.Close(t, writer1)
		// when
		writer2, err := log.New(dir).OpenWriter()
		defer tests.Close(t, writer2)
		// then
		assert.ErrorIs(t, err, log.ErrLocked)
	})
}

func TestLog_Segments(t *testing.T) {
	t.Run("new log should have no segments", func(t *testing.T) {
		dir := tests.TempDir(t)
		l := log.New(dir)
		// when
		segments, err := l.Segments()
		// then
		require.NoError(t, err)
		assert.Empty(t, segments)
	})

	t.Run("should create a new segment on first write", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t)
		t1, _ := writer.Write(data1)
		// when
		segments, err := l.Segments()
		// then
		require.NoError(t, err)
		require.Len(t, segments, 1)
		assert.Len(t, segments, 1)
		assert.True(t,
			segments[0].StartingAt.Before(t1) || segments[0].StartingAt.Equal(t1),
			"segment start should be <= time of first of entry")
	})

	t.Run("should return sorted segments", func(t *testing.T) {
		entry := make([]byte, tests.OneMegabyte)
		startingTime := time2006

		testCases := map[string][]log.OpenWriterOption{
			"real time": {log.MaxSegmentSizeMB(1)},
			"fixed now": {log.MaxSegmentSizeMB(1), log.NowFunc(fixedNow(startingTime))},
		}

		for name, options := range testCases {
			t.Run(name, func(t *testing.T) {
				l, writer := tests.OpenLogWithWriter(t, options...)
				for i := 0; i < 5; i++ {
					_, err := writer.Write(entry)
					require.NoError(t, err)
				}
				// when
				segments, err := l.Segments()
				// then
				require.NoError(t, err)
				for i := 0; i < len(segments)-1; i++ {
					assert.Truef(t, segments[i].StartingAt.Before(segments[i+1].StartingAt),
						"segment[%d]=%s is not before segment[%d]=%s", i, segments[i].StartingAt, i+1, segments[i+1].StartingAt)
				}
			})
		}
	})
}

func TestLog_RemoveSegment(t *testing.T) {
	t.Run("should return error when segment does not exist", func(t *testing.T) {
		l := log.New(tests.TempDir(t))
		err := l.RemoveSegmentStartingAt(time.Now())
		assert.Error(t, err)
	})

	t.Run("should not be possible to remove last segment", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t)
		tests.WriteEntry(t, writer, tests.OneMegabyte)
		segments, _ := l.Segments()
		assert.Len(t, segments, 1)
		// when
		err := l.RemoveSegmentStartingAt(segments[0].StartingAt)
		// then
		assert.ErrorIs(t, err, log.ErrInvalidParameter)
	})

	t.Run("should remove segment", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t, log.MaxSegmentSizeMB(1))
		tests.WriteEntry(t, writer, tests.OneMegabyte)
		tests.WriteEntry(t, writer, tests.OneMegabyte)
		segmentsBefore, err := l.Segments()
		require.NoError(t, err)
		// when
		err = l.RemoveSegmentStartingAt(segmentsBefore[0].StartingAt)
		// then
		require.NoError(t, err)
		segmentsAfter, err := l.Segments()
		require.NoError(t, err)
		assert.Equal(t, segmentsBefore[1:], segmentsAfter)
	})
}

func TestLog_LastEntry(t *testing.T) {
	t.Run("should return error when log is empty", func(t *testing.T) {
		l := log.New(tests.TempDir(t))
		// when
		_, bytes, err := l.LastEntry()
		// then
		assert.ErrorIs(t, err, log.ErrEOL)
		assert.Nil(t, bytes)
	})

	t.Run("should return last entry for single entry log", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t)
		t1, _ := writer.Write(data1)
		// when
		actualTime, bytes, err := l.LastEntry()
		// then
		require.NoError(t, err)
		assert.Equal(t, data1, bytes)
		assert.True(t, t1.Equal(actualTime))
	})

	t.Run("should return last entry for two entries log", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t)
		_, _ = writer.Write(data1)
		t2, _ := writer.Write(data2)
		// when
		actualTime, bytes, err := l.LastEntry()
		// then
		require.NoError(t, err)
		assert.Equal(t, data2, bytes)
		assert.True(t, t2.Equal(actualTime))
	})
}

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time {
		return t
	}
}
