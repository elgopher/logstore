// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gofrs/flock"
)

func (l *Log) openWriter(options []OpenWriterOption) (Writer, error) {
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

	filename, err := l.segmentFilenameForWriter(settings.now)
	if err != nil {
		return nil, err
	}

	file := path.Join(l.dir, filename)

	segmentFile, err := openFileForAppending(file)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("stat failed for file %s: %w", file, err)
	}

	size := stat.Size()

	return &writer{
		currentSegment: &segmentWriter{
			file:      segmentFile,
			sizeBytes: size,
			startTime: segmentFilename(filename).StartedAt(),
		},
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

type writer struct {
	currentSegment      *segmentWriter
	now                 func() time.Time
	maxSegmentSizeBytes int64
	maxSegmentDuration  time.Duration
	lastTime            time.Time
	lock                *flock.Flock
	dir                 string
}

type segmentWriter struct {
	file      *os.File
	sizeBytes int64
	startTime time.Time
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
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("closing segment file failed: %w", err)
	}

	return nil
}

func (w *writer) Close() error {
	if err := w.lock.Unlock(); err != nil {
		_ = w.currentSegment.close()

		return fmt.Errorf("error unlocking the log lock: %w", err)
	}

	if err := w.currentSegment.close(); err != nil {
		return fmt.Errorf("closing Writer failed: %w", err)
	}

	return nil
}

func (w *writer) Write(entry []byte, options ...WriteOption) (time.Time, error) {
	t := w.now()

	if !t.After(w.lastTime) {
		t = w.lastTime.Add(time.Nanosecond)
	}

	if err := w.writeEntry(t, entry); err != nil {
		return time.Time{}, err
	}

	w.lastTime = t

	return t, nil
}

func (w *writer) writeEntry(t time.Time, entry []byte) error {
	timeBinary, err := t.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshaling entry time failed: %w", err)
	}

	if _, err = w.currentSegment.Write(timeBinary); err != nil {
		return fmt.Errorf("writing entry time failed: %w", err)
	}

	if err = binary.Write(w.currentSegment, binary.LittleEndian, uint32(len(entry))); err != nil {
		return fmt.Errorf("writing entry len failed: %w", err)
	}

	if _, err = w.currentSegment.Write(entry); err != nil {
		return fmt.Errorf("writing entry data failed: %w", err)
	}

	if w.currentSegment.maxSizeExceeded(w.maxSegmentSizeBytes) ||
		w.currentSegment.maxDurationExceeded(t, w.maxSegmentDuration) {
		if err := w.rollOver(t.Add(time.Nanosecond)); err != nil {
			return err
		}
	}

	return nil
}

func (w *writer) rollOver(start time.Time) error {
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
