// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"io"
	"os"
	"time"
)

func Open(dir string) (*Log, error) {
	f, err := os.OpenFile(dir+"/data", os.O_RDWR, 0664)
	if err != nil {
		return nil, err
	}
	return &Log{file: f}, nil
}

type Log struct {
	file *os.File
}

func (l *Log) Append(b []byte, o ...AppendOption) (time.Time, error) {
	_, err := l.file.Write(b)
	if err != nil {
		//	go backward by n?
		return time.Time{}, err
	}
	return time.Time{}, nil
}

type AppendOption func()

// t must be greater than the tail of log
func ForceTime(t time.Time) AppendOption {
	return func() {

	}
}

func (l *Log) Close() error {
	return nil
}

func (l *Log) Reader(options ...ReaderOption) Reader {
	return nil
}

type ReaderOption func()

func StartTime(time.Time) ReaderOption {
	return nil
}

func StopTime(time.Time) ReaderOption {
	return nil
}

var DontReportEol ReaderOption = func() {

}

func IsEOL(err error) bool {
	return false
}

type Reader interface {
	Read(...ReadOption) (time.Time, []byte, error)
	io.Closer
}

type reader struct {
}

// When error is reported, next call will try to read same record
func (r *reader) Read(...ReadOption) (time.Time, []byte, error) {
	return time.Now(), make([]byte, 4096), nil
}

type ReadOption func()

// AppendToBuffer can be used for optimization purposes - when it is passed to Read, then this buffer
// is used as output instead of allocating a new one.
func AppendToBuffer(buffer []byte) ReadOption {
	return func() {
	}
}

func (r *reader) Close() error {
	return nil
}

// Entry can be event, transaction etc.
type Entry struct {
	Time time.Time
	Data []byte
}

func (l *Log) Segments() ([]Segment, error) {
	return nil, nil
}

type Segment struct {
	Start time.Time
	Stop  time.Time // stop is deducted from next segment when calling Log.Segments
}

func (l *Log) RemoveSegmentStartingAt(start time.Time) error {
	return nil
}
