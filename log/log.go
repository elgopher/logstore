// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
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

func (l *Log) OpenWriter(options ...OpenWriterOption) (*Writer, error) {
	return l.openWriter(options)
}

type OpenWriterOption func(*WriterSettings) error

type WriterSettings struct {
	now                 func() time.Time
	maxSegmentSizeBytes int64
	maxSegmentDuration  time.Duration
}

func NowFunc(f func() time.Time) OpenWriterOption {
	return func(s *WriterSettings) error {
		s.now = f

		return nil
	}
}

func MaxSegmentSizeMB(megabytes int) OpenWriterOption {
	return func(s *WriterSettings) error {
		s.maxSegmentSizeBytes = int64(megabytes) * oneMegabyte

		return nil
	}
}

func MaxSegmentDuration(duration time.Duration) OpenWriterOption {
	return func(s *WriterSettings) error {
		s.maxSegmentDuration = duration

		return nil
	}
}

func (l *Log) OpenReader(options ...OpenReaderOption) (Reader, error) {
	return l.openReader(options)
}

type OpenReaderOption func(*ReaderSettings) error

type ReaderSettings struct {
	openOldestSegment func(dir string, segments []Segment) (*os.File, int, error)
}

func StartingFrom(t time.Time) OpenReaderOption {
	return func(s *ReaderSettings) error {
		s.openOldestSegment = func(dir string, segments []Segment) (*os.File, int, error) {
			return openSegmentStartingAt(t, dir, segments)
		}

		return nil
	}
}

type Reader interface {
	Read() (time.Time, []byte, error)
	Close() error
}

func (l *Log) Segments() ([]Segment, error) {
	var segments []Segment

	files, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("os.ReadDir failed: %w", err)
	}

	for _, f := range files {
		name := f.Name()
		if !f.IsDir() && strings.HasSuffix(name, ".segment") {
			file := segmentFilename(name)
			segment := Segment{StartingAt: file.StartedAt()}
			segments = append(segments, segment)
		}
	}

	return segments, nil
}

func (l *Log) RemoveSegmentStartingAt(t time.Time) error {
	segmentFilename := path.Join(l.dir, segmentFilenameStartingAt(t))

	segments, err := l.Segments()
	if err != nil {
		return fmt.Errorf("listing segments failed: %w", err)
	}

	if len(segments) == 1 {
		return fmt.Errorf("cant remove last segment: %w", ErrInvalidParameter)
	}

	if err := os.Remove(segmentFilename); err != nil {
		return fmt.Errorf("removing file %s failed %w", segmentFilename, err)
	}

	return nil
}

func (l *Log) LastEntry() (time.Time, []byte, error) {
	reader, err := l.OpenReader()
	if err != nil {
		return time.Time{}, nil, err
	}

	defer func() {
		_ = reader.Close()
	}()

	var (
		lastTime   time.Time
		lastData   []byte
		entryFound bool
	)

	for {
		t, data, err := reader.Read()
		if errors.Is(err, ErrEOL) {
			if !entryFound {
				return time.Time{}, nil, fmt.Errorf("log is empty: %w", ErrEOL)
			}

			return lastTime, lastData, nil
		}

		if err != nil {
			return time.Time{}, nil, fmt.Errorf("error reading last entry time from segment file: %w", err)
		}

		lastTime = t
		lastData = data
		entryFound = true
	}
}

type Segment struct {
	StartingAt time.Time
}
