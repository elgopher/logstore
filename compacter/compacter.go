// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package compacter

import (
	"context"
	"fmt"
	stdlog "log"
	"time"

	"github.com/jacekolszak/logstore/log"
)

func RemoveOldSegments(l Log, olderThan time.Time) (Results, error) {
	segments, err := l.Segments()
	if err != nil {
		return Results{}, fmt.Errorf("listing segments failed: %w", err)
	}

	res := Results{}

	for _, segment := range segments {
		if segment.StartingAt.Before(olderThan) {
			if err := l.RemoveSegmentStartingAt(segment.StartingAt); err != nil {
				return res, fmt.Errorf("removing segment failed: %w", err)
			}

			res.SegmentsRemoved = append(res.SegmentsRemoved, segment)
		}
	}

	return res, nil
}

type Log interface {
	Segments() ([]log.Segment, error)
	RemoveSegmentStartingAt(t time.Time) error
}

type Results struct {
	SegmentsRemoved []log.Segment
}

func Start(ctx context.Context, l Log, options ...Option) error {
	if l == nil {
		return fmt.Errorf("nil log: %w", log.ErrInvalidParameter)
	}

	const sevenDays = time.Hour * 24 * 7

	settings := &Settings{
		interval:  time.Hour,
		retention: sevenDays,
	}

	for _, applyOption := range options {
		if err := applyOption(settings); err != nil {
			return fmt.Errorf("error applying option: %w", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(settings.interval):
			olderThan := time.Now().Add(-settings.retention)

			results, err := RemoveOldSegments(l, olderThan)
			if err != nil {
				stdlog.Printf("compacter.RemoveOldSegments failed: %s", err)
			} else {
				count := len(results.SegmentsRemoved)
				if count > 0 {
					stdlog.Printf("%d segments removed older than %s", count, olderThan)
				}
			}
		}
	}
}

type Option func(*Settings) error

type Settings struct {
	interval  time.Duration
	retention time.Duration
}

func Interval(duration time.Duration) Option {
	return func(s *Settings) error {
		s.interval = duration

		return nil
	}
}

func Retention(duration time.Duration) Option {
	return func(s *Settings) error {
		s.retention = duration

		return nil
	}
}
