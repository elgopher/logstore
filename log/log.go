// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"time"
)

func New(dir string) *Log {
	return &Log{
		dir: dir,
	}
}

type Log struct {
	dir string
}

func (l *Log) OpenWriter(options ...OpenWriterOption) (Writer, error) {
	return l.openWriter(options)
}

type OpenWriterOption func(*WriterSettings) error

type WriterSettings struct {
	now func() time.Time
}

func NowFunc(f func() time.Time) OpenWriterOption {
	return func(s *WriterSettings) error {
		s.now = f

		return nil
	}
}

type Writer interface {
	Write(entry []byte, options ...WriteOption) (time.Time, error)
	Close() error
}

type WriteOption func() error

func (l *Log) OpenReader(options ...OpenReaderOption) (Reader, error) {
	return l.openReader(options)
}

type OpenReaderOption func(*ReaderSettings) error

type ReaderSettings struct{}

type Reader interface {
	Read() (time.Time, []byte, error)
	Close() error
}
