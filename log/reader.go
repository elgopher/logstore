// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

func (l *Log) openReader(options []OpenReaderOption) (Reader, error) {
	settings := &ReaderSettings{
		openOldestSegment: openOldestSegmentAtTheBegging,
	}

	for _, applyOption := range options {
		if applyOption == nil {
			continue
		}

		if err := applyOption(settings); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	segments, err := l.Segments()
	if err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return &emptyLogReader{}, nil
	}

	segmentFile, segmentIndex, err := settings.openOldestSegment(l.dir, segments)
	if err != nil {
		return nil, err
	}

	return &segmentsReader{
		segmentFile:    segmentFile,
		segments:       segments,
		currentSegment: segmentIndex,
		dir:            l.dir,
	}, nil
}

func openOldestSegmentAtTheBegging(dir string, segments []Segment) (*os.File, int, error) {
	const oldestSegmentIndex = 0
	oldestSegment := segments[oldestSegmentIndex]

	f, err := openSegmentFileForRead(dir, oldestSegment)
	if err != nil {
		return nil, 0, err
	}

	return f, oldestSegmentIndex, nil
}

func openSegmentStartingAt(t time.Time, dir string, segments []Segment) (*os.File, int, error) {
	oldestSegmentIndex := 0

	for i, segment := range segments {
		if segment.StartingAt.After(t) {
			break
		}

		oldestSegmentIndex = i
	}

	oldestSegment := segments[oldestSegmentIndex]

	f, err := openSegmentFileForRead(dir, oldestSegment)
	if err != nil {
		return nil, 0, err
	}

	pos, err := findClosestEntryPosition(t, f)
	if err != nil {
		return nil, 0, err
	}

	if _, err = f.Seek(pos, io.SeekStart); err != nil {
		return nil, 0, fmt.Errorf("seeking to entry starting position failed: %w", err)
	}

	return f, oldestSegmentIndex, nil
}

type emptyLogReader struct{}

func (r *emptyLogReader) Read() (time.Time, []byte, error) {
	return time.Time{}, nil, ErrEOL
}

func (r *emptyLogReader) Close() error {
	return nil
}

func openSegmentFileForRead(dir string, segment Segment) (*os.File, error) {
	f, err := os.Open(path.Join(dir, segmentFilenameStartingAt(segment.StartingAt)))
	if err != nil {
		return nil, fmt.Errorf("opening segment file failed: %w", err)
	}

	return f, nil
}

type segmentsReader struct {
	segmentFile    *os.File
	segments       []Segment
	currentSegment int
	dir            string
}

func (r *segmentsReader) Read() (time.Time, []byte, error) {
	t, data, err := decodeEntry(r.segmentFile)
	if errors.Is(err, io.EOF) {
		r.currentSegment++
		if r.currentSegment >= len(r.segments) {
			return time.Time{}, nil, ErrEOL
		}

		_ = r.segmentFile.Close()

		r.segmentFile, err = openSegmentFileForRead(r.dir, r.segments[r.currentSegment])
		if err != nil {
			return time.Time{}, nil, err
		}

		return r.Read()
	}

	return t, data, nil
}

func (r *segmentsReader) Close() error {
	if err := r.segmentFile.Close(); err != nil {
		return fmt.Errorf("error closing segment file: %w", err)
	}

	return nil
}

func (l *Log) readLastTime() (time.Time, error) {
	reader, err := l.OpenReader()
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		_ = reader.Close()
	}()

	var lastTime time.Time

	for {
		t, _, err := reader.Read()
		if errors.Is(err, ErrEOL) {
			return lastTime, nil
		}

		if err != nil {
			return time.Time{}, fmt.Errorf("error reading last entry time from segment file: %w", err)
		}

		lastTime = t
	}
}
