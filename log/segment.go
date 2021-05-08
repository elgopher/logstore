// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

const (
	segmentFilenameDateFormat = "2006-01-02T15_04_05.000000000Z"
	segmentFilenameExtension  = ".segment"
)

type segmentFilename string

func (s segmentFilename) StartedAt() time.Time {
	timeString := strings.TrimSuffix(string(s), segmentFilenameExtension)

	t, err := time.Parse(segmentFilenameDateFormat, timeString)
	if err != nil {
		panic(err)
	}

	return t
}

func segmentFilenameStartingAt(t time.Time) string {
	return t.UTC().Format(segmentFilenameDateFormat) + segmentFilenameExtension
}

type segmentWriter struct {
	file      *os.File
	sizeBytes int64
	startTime time.Time
}

func (l *Log) openLastUsedSegmentWriter() (*segmentWriter, error) {
	segments, err := l.Segments()
	if err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return nil, nil
	}

	lastSegment := segments[len(segments)-1]

	return openSegmentWriter(l.dir, lastSegment.StartingAt)
}

func openSegmentWriter(dir string, startTime time.Time) (*segmentWriter, error) {
	filename := path.Join(dir, segmentFilenameStartingAt(startTime))

	segmentFile, err := openFileForAppending(filename)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("stat failed for file %s: %w", filename, err)
	}

	size := stat.Size()

	return &segmentWriter{
		file:      segmentFile,
		sizeBytes: size,
		startTime: startTime,
	}, nil
}

func (l *segmentWriter) Write(b []byte) (int, error) {
	n, err := l.file.Write(b)
	defer func() {
		l.sizeBytes += int64(n)
	}()

	if err != nil {
		return n, fmt.Errorf("writing to segment failed: %w", err)
	}

	return n, nil
}

func (l *segmentWriter) maxSizeExceeded(maxSize int64) bool {
	return l.sizeBytes > maxSize
}

func (l *segmentWriter) maxDurationExceeded(t time.Time, maxSegmentDuration time.Duration) bool {
	return t.After(l.startTime.Add(maxSegmentDuration))
}

func (l *segmentWriter) close() error {
	if l == nil {
		return nil
	}

	if err := l.file.Close(); err != nil {
		return fmt.Errorf("closing segment file failed: %w", err)
	}

	return nil
}
