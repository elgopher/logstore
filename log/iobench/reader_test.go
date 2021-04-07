package iobench

import (
	"bytes"
	"testing"
)

func BenchmarkSlowReader(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 4096))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := OpenSlowReader()
		for j := 0; j < 100; j++ {
			buffer.Reset()
			r.Read(buffer)
		}
		r.Close() // 58us/op
	}
}

func BenchmarkFastReader(b *testing.B) {
	buffer := make([]byte, 4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := OpenFastreader()
		for j := 0; j < 100; j++ {
			r.Read(buffer)
		}
		r.Close() // 53 us/op
	}
}
