// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log_test

import (
	"testing"
	"time"

	"github.com/jacekolszak/logstore/internal/tests"
	"github.com/stretchr/testify/require"
)

func BenchmarkWriter_WriteWithTime(b *testing.B) {
	b.Run("should not make any allocations", func(b *testing.B) {
		b.ReportAllocs()

		writer := tests.OpenLogWriter(b)
		t := time.Time{}

		for i := 0; i < b.N; i++ {
			t = t.Add(time.Nanosecond)

			err := writer.WriteWithTime(t, data1) // still 2 allocations in the codec.go (should be fixed soon)
			require.NoError(b, err)
		}
	})
}
