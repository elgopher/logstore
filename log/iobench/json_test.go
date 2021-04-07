package iobench

import (
	"testing"

	"github.com/jacekolszak/logstore/codec"
	"github.com/jacekolszak/logstore/log"
)

func BenchmarkAppend(b *testing.B) {
	b.ReportAllocs()
	l, err := log.Open("/tmp")
	if err != nil {
		panic(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = codec.Append(l, stubEncoder)
		if err != nil {
			panic(err)
		}
	}
}

func stubEncoder() ([]byte, error) {
	return make([]byte, 100), nil
}
