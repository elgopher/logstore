// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package compacter

import (
	"time"

	"github.com/jacekolszak/logstore/log"
)

func RemoveOldSegments(l *log.Log, olderThan time.Time) (Results, error) {
	segments, err := l.Segments()
	if err != nil {
		return Results{}, err
	}
	for _, segment := range segments {
		if segment.Stop.Before(olderThan) {
			if err = l.RemoveSegmentStartingAt(segment.Start); err != nil {
				panic(err)
			}
		}
	}
	return Results{}, nil
}

type Results struct {
	SegmentsRemoved []log.Segment
}
