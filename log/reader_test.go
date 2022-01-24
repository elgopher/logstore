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

func TestReader_Read(t *testing.T) {
	t.Run("should return ErrEOL when no entries were written before", func(t *testing.T) {
		reader := tests.OpenLogReader(t)
		_, data, err := reader.Read()
		assert.ErrorIs(t, err, log.ErrEOL)
		assert.Nil(t, data)
	})

	t.Run("should read two entries written using two writers", func(t *testing.T) {
		l, writer1 := tests.OpenLogWithWriter(t)
		t1, _ := writer1.Write(data1)
		_ = writer1.Close()
		writer2, _ := l.OpenWriter()
		t2, err := writer2.Write(data1)
		defer tests.Close(t, writer2)
		require.NoError(t, err)
		// when
		entries := tests.ReadAll(t, l)
		// then
		assert.Len(t, entries, 2)
		assert.True(t, entries[0].Time.Equal(t1))
		assert.True(t, entries[1].Time.Equal(t2))
	})

	t.Run("should read entries starting from given time", func(t *testing.T) {
		t.Run("when given time is before than first entry", func(t *testing.T) {
			firstEntryTime := time2006

			l, writer := tests.OpenLogWithWriter(t, log.NowFunc(fixedNow(firstEntryTime)))
			t1, _ := writer.Write(data1)

			givenTime := time2005
			// when
			actual := tests.ReadAll(t, l, log.StartingFrom(givenTime))
			// then
			require.Len(t, actual, 1)
			assert.True(t, t1.Equal(actual[0].Time))
			assert.Equal(t, data1, actual[0].Data)
		})

		t.Run("when given time is after the last entry", func(t *testing.T) {
			firstEntryTime := time2005

			l, writer := tests.OpenLogWithWriter(t, log.NowFunc(fixedNow(firstEntryTime)))
			_, _ = writer.Write(data1)

			givenTime := time2006
			// when
			actual := tests.ReadAll(t, l, log.StartingFrom(givenTime))
			// then
			require.Len(t, actual, 0)
		})

		t.Run("when entry is at the end of the sole segment", func(t *testing.T) {
			l, writer := tests.OpenLogWithWriter(t)
			_, _ = writer.Write(data1)
			t2, _ := writer.Write(data2)
			// when
			actual := tests.ReadAll(t, l, log.StartingFrom(t2))
			// then
			require.Len(t, actual, 1)
			assert.True(t, t2.Equal(actual[0].Time))
			assert.Equal(t, data2, actual[0].Data)
		})

		t.Run("when given time is between two entries in a sole segment", func(t *testing.T) {
			currentTime := time.Time{}
			clock := tests.Clock{CurrentTime: &currentTime}

			l, writer := tests.OpenLogWithWriter(t, log.NowFunc(clock.Now))
			t1, _ := writer.Write(data1)
			clock.MoveForward(time.Hour)
			t2, _ := writer.Write(data2)
			// when
			afterT1beforeT2 := t1.Add(time.Minute)
			actual := tests.ReadAll(t, l, log.StartingFrom(afterT1beforeT2))
			// then
			require.Len(t, actual, 1)
			assert.True(t, t2.Equal(actual[0].Time))
			assert.Equal(t, data2, actual[0].Data)
		})

		t.Run("when entry is at the beginning of second segment", func(t *testing.T) {
			currentTime := time.Time{}
			clock := tests.Clock{CurrentTime: &currentTime}

			l, writer := tests.OpenLogWithWriter(t, log.NowFunc(clock.Now), log.MaxSegmentDuration(time.Minute))
			clock.MoveForward(time.Hour)
			_, _ = writer.Write(data1)
			t2, _ := writer.Write(data2)
			// when
			actual := tests.ReadAll(t, l, log.StartingFrom(t2))
			// then
			require.Len(t, actual, 1)
			assert.True(t, t2.Equal(actual[0].Time))
			assert.Equal(t, data2, actual[0].Data)
		})
	})
}
