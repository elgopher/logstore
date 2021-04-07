package iobench

import "testing"

func BenchmarkRead(b *testing.B) {
	b.ReportAllocs()
	out := make([]byte, 4096)
	for i := 0; i < b.N; i++ {
		_ = Read(out)
		out[0] = '1'
	}
}

func BenchmarkReadWithReturn(b *testing.B) {
	b.ReportAllocs()
	res := byte(0)
	for i := 0; i < b.N; i++ {
		r := ReadWithReturn()
		res += r[0] + r[1] + r[2] + r[3] + r[4]
	}
}
