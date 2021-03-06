// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/elgopher/logstore/log"
	"github.com/stretchr/testify/require"
)

func OpenLogWriter(t TestingT, options ...log.OpenWriterOption) *log.Writer {
	t.Helper()

	_, writer := OpenLogWithWriter(t, options...)

	return writer
}

func OpenLogWithWriter(t TestingT, options ...log.OpenWriterOption) (*log.Log, *log.Writer) {
	t.Helper()

	l := log.New(TempDir(t))
	writer, err := l.OpenWriter(options...)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = writer.Close()
	})

	return l, writer
}

func OpenLogReader(t *testing.T, options ...log.OpenReaderOption) log.Reader {
	t.Helper()

	l := log.New(TempDir(t))
	reader, err := l.OpenReader(options...)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = reader.Close()
	})

	return reader
}

func OpenReader(t TestingT, l *log.Log, options ...log.OpenReaderOption) log.Reader {
	t.Helper()

	reader, err := l.OpenReader(options...)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = reader.Close()
	})

	return reader
}

func ReadAll(t *testing.T, l *log.Log, options ...log.OpenReaderOption) []Entry {
	t.Helper()

	reader, err := l.OpenReader(options...)
	defer Close(t, reader)
	require.NoError(t, err)

	var entries []Entry

	for {
		v, data, err := reader.Read()
		if errors.Is(err, log.ErrEOL) {
			return entries
		}

		require.NoError(t, err)

		entries = append(entries, Entry{Time: v, Data: data})
	}
}

type Entry struct {
	Time time.Time
	Data []byte
}

func WriteEntry(t *testing.T, writer *log.Writer, sizeInBytes int64) time.Time {
	t.Helper()

	data := make([]byte, sizeInBytes)
	entryTime, err := writer.Write(data)
	require.NoError(t, err)

	return entryTime
}
