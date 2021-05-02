// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
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

func (l *Log) segmentFilenameForWriter(now func() time.Time) (string, error) {
	segments, err := l.Segments()
	if err != nil {
		return "", err
	}

	var filename string
	if len(segments) == 0 {
		filename = segmentFilenameStartingAt(now())
	} else {
		lastSegment := segments[len(segments)-1]
		filename = segmentFilenameStartingAt(lastSegment.StartingAt)
	}

	return filename, nil
}
