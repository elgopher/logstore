// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package compacter_test

import (
	"context"
	"testing"
	"time"

	"github.com/jacekolszak/logstore/compacter"
	"github.com/jacekolszak/logstore/internal/tests"
	"github.com/jacekolszak/logstore/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveOldSegments(t *testing.T) {
	t.Run("should remove old segments", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t, log.MaxSegmentSizeMB(1))
		for i := 0; i < 10; i++ {
			tests.WriteEntry(t, writer, tests.OneMegabyte)
		}
		segmentsBefore, err := l.Segments()
		require.NoError(t, err)
		// when
		results, err := compacter.RemoveOldSegments(l, segmentsBefore[4].StartingAt)
		// then
		require.NoError(t, err)
		segmentsAfter, err := l.Segments()
		require.NoError(t, err)
		assert.Equal(t, results.SegmentsRemoved, segmentsBefore[:4])
		assert.Equal(t, segmentsAfter, segmentsBefore[4:])
	})
}

func TestStart(t *testing.T) {
	t.Run("should return error for nil log", func(t *testing.T) {
		err := compacter.Start(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("should stop once context is cancelled", func(t *testing.T) {
		l := log.New(tests.TempDir(t))
		ctx, cancel := context.WithCancel(context.Background())
		var err error
		async := tests.RunAsync(func() {
			err = compacter.Start(ctx, l)
		})
		// when
		cancel()
		// then
		async.WaitOrFailAfter(t, time.Second)
		assert.NoError(t, err)
	})

	t.Run("should skip nil option", func(t *testing.T) {
		l := log.New(tests.TempDir(t))
		ctx, cancel := context.WithCancel(context.Background())
		var err error
		async := tests.RunAsync(func() {
			err = compacter.Start(ctx, l)
		})
		cancel()
		async.WaitOrFailAfter(t, time.Second)
		assert.NoError(t, err)
	})

	t.Run("should return error when option returned error", func(t *testing.T) {
		l := log.New(tests.TempDir(t))
		option := func(s *compacter.Settings) error {
			return tests.ErrFixed
		}
		// when
		err := compacter.Start(context.Background(), l, option)
		// then
		assert.Error(t, err)
	})

	t.Run("should continuously compact versions in the background", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t, log.MaxSegmentSizeMB(1))
		ctx, cancel := context.WithCancel(context.Background())
		async := tests.RunAsync(func() {
			_ = compacter.Start(ctx, l, compacter.Interval(time.Millisecond), compacter.Retention(time.Millisecond))
		})
		// when
		tests.WriteEntry(t, writer, tests.OneMegabyte)
		tests.WriteEntry(t, writer, tests.OneMegabyte)
		tests.WriteEntry(t, writer, tests.OneMegabyte)
		// then
		assert.Eventually(t, numberOfSegments(l, 1), 100*time.Millisecond, time.Millisecond)
		// and when
		tests.WriteEntry(t, writer, tests.OneMegabyte)
		// then
		assert.Eventually(t, numberOfSegments(l, 1), 100*time.Millisecond, time.Millisecond)
		// cleanup
		cancel()
		async.WaitOrFailAfter(t, time.Second)
	})
}

func numberOfSegments(l *log.Log, expected int) func() bool {
	return func() bool {
		segments, err := l.Segments()
		if err != nil {
			panic(err)
		}

		return len(segments) == expected
	}
}
