package iobench

import (
	"bytes"
	"io"
	"os"
)

func OpenSlowReader() *SlowReader {
	f, err := os.Open("/tmp/dupa")
	if err != nil {
		panic(err)
	}
	return &SlowReader{file: f, buffer: bytes.NewBuffer(make([]byte, 4096))}
}

func OpenFastreader() *FastReader {
	f, err := os.Open("/tmp/dupa")
	if err != nil {
		panic(err)
	}
	return &FastReader{file: f}
}

type SlowReader struct {
	file   *os.File
	buffer *bytes.Buffer
}

func (r *SlowReader) Read(writer io.Writer) {
	r.buffer.Reset()
	b := r.buffer.Bytes()
	n, err := r.file.Read(b[0:cap(b)])
	if err != nil {
		panic(err)
	}
	writer.Write(b[:n])
}

func (r *SlowReader) Close() {
	r.file.Close()
}

type FastReader struct {
	file *os.File
}

func (r *FastReader) Read(b []byte) {
	_, err := r.file.Read(b)
	if err != nil {
		panic(err)
	}
}

func (r *FastReader) Close() {
	r.file.Close()
}
