// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log_test

import (
	"testing"

	"github.com/jacekolszak/logstore/internal/tests"
	"github.com/jacekolszak/logstore/log"
	"github.com/stretchr/testify/assert"
)

func TestReader_Read(t *testing.T) {
	t.Run("should return ErrEOL when no entries were written before", func(t *testing.T) {
		reader := tests.OpenLogReader(t)
		_, data, err := reader.Read()
		assert.ErrorIs(t, err, log.ErrEOL)
		assert.Nil(t, data)
	})
}
