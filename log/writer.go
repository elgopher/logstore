// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gofrs/flock"
)

func (l *Log) openWriter(options []OpenWriterOption) (*Writer, error) {
	settings, err := l.writerSettings(options)
	if err != nil {
		return nil, err
	}

	if err := mkdirIfMissing(l.dir); err != nil {
		return nil, err
	}

	lock, err := tryLock(l.dir)
	if err != nil {
		return nil, err
	}

	lastTime, err := l.readLastTime()
	if err != nil {
		return nil, err
	}

	currentSegment, err := l.openLastUsedSegmentWriter()
	if err != nil {
		return nil, err
	}

	return &Writer{
		currentSegment:      currentSegment,
		now:                 settings.now,
		lastTime:            lastTime,
		maxSegmentSizeBytes: settings.maxSegmentSizeBytes,
		maxSegmentDuration:  settings.maxSegmentDuration,
		lock:                lock,
		dir:                 l.dir,
	}, nil
}

func (l *Log) writerSettings(options []OpenWriterOption) (*WriterSettings, error) {
	settings := &WriterSettings{
		now:                 time.Now,
		maxSegmentSizeBytes: oneGigabyte,
		maxSegmentDuration:  oneMonth,
	}

	for _, applyOption := range options {
		if applyOption == nil {
			continue
		}

		if err := applyOption(settings); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	return settings, nil
}

func openFileForAppending(file string) (*os.File, error) {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664)
	if err != nil {
		return nil, fmt.Errorf("error opening segment file %s for write: %w", file, err)
	}

	return f, nil
}

func mkdirIfMissing(dir string) error {
	_, err := os.Stat(path.Join(dir))
	if os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0775); err != nil {
			return fmt.Errorf("cannot create directory: %w", err)
		}
	}

	return nil
}

func tryLock(dir string) (*flock.Flock, error) {
	lockFile := path.Join(dir, "log.lock")
	lock := flock.New(lockFile)

	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("error trying to lock log for writing: %w", err)
	}

	if !locked {
		return nil, ErrLocked
	}

	return lock, nil
}

type Writer struct {
	currentSegment      *segmentWriter
	now                 func() time.Time
	maxSegmentSizeBytes int64
	maxSegmentDuration  time.Duration
	lastTime            time.Time
	lock                *flock.Flock
	dir                 string
}

func (w *Writer) Close() error {
	if err := w.lock.Unlock(); err != nil {
		_ = w.currentSegment.close()

		return fmt.Errorf("error unlocking the log lock: %w", err)
	}

	if err := w.currentSegment.close(); err != nil {
		return fmt.Errorf("closing Writer failed: %w", err)
	}

	return nil
}

func (w *Writer) Write(entry []byte) (time.Time, error) {
	t := w.now()

	if !t.After(w.lastTime) {
		t = w.lastTime.Add(time.Nanosecond)
	}

	return t, w.WriteWithTime(t, entry)
}

func (w *Writer) WriteWithTime(t time.Time, entry []byte) error {
	if !t.After(w.lastTime) {
		return fmt.Errorf("forced time is not after last entry time: %w", ErrInvalidParameter)
	}

	if err := w.writeEntry(t, entry); err != nil {
		return err
	}

	w.lastTime = t

	return nil
}

func (w *Writer) writeEntry(t time.Time, entry []byte) error {
	if w.currentSegment == nil {
		var err error

		w.currentSegment, err = openSegmentWriter(w.dir, t)
		if err != nil {
			return err
		}
	}

	if err := encodeEntry(w.currentSegment, t, entry); err != nil {
		return err
	}

	if w.currentSegment.maxSizeExceeded(w.maxSegmentSizeBytes) ||
		w.currentSegment.maxDurationExceeded(t, w.maxSegmentDuration) {
		if err := w.rollOver(t.Add(time.Nanosecond)); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) rollOver(start time.Time) error {
	if err := w.currentSegment.close(); err != nil {
		return fmt.Errorf("error closing segment file: %w", err)
	}

	f := path.Join(w.dir, segmentFilenameStartingAt(start))

	segmentFile, err := openFileForAppending(f)
	if err != nil {
		return err
	}

	w.currentSegment = &segmentWriter{
		file:      segmentFile,
		sizeBytes: 0,
		startTime: start,
	}

	return nil
}
