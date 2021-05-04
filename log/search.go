package log

import (
	"errors"
	"fmt"
	"io"
	"time"
)

func findClosestEntryPosition(t time.Time, file io.ReadSeeker) (int64, error) {
	for {
		// Following algorithm is extremely inefficient and will be replaced with binary search soon
		entryStartingPosition, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			return 0, fmt.Errorf("getting file position failed: %w", err)
		}

		entryTime, _, err := decodeEntry(file)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return entryStartingPosition, nil
			}

			return 0, err
		}

		if !entryTime.Before(t) {
			return entryStartingPosition, nil
		}
	}
}
