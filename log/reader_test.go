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
}
