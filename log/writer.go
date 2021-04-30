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

	currentSegment, err := openFileForAppending(file)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("stat failed for file %s: %w", file, err)
	}

	size := stat.Size()

	return &writer{
		now:              settings.now,
		lastTime:         lastTime,
		currentSegment:   currentSegment,
		currentSizeBytes: size,
		maxSizeBytes:     settings.maxSegmentSizeBytes,
		lock:             lock,
		dir:              l.dir,
	}, nil
}

func (l *Log) writerSettings(options []OpenWriterOption) (*WriterSettings, error) {
	settings := &WriterSettings{
		now:                 time.Now,
		maxSegmentSizeBytes: oneGigabyte,
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
	now              func() time.Time
	currentSegment   *os.File
	currentSizeBytes int64
	maxSizeBytes     int64
	lastTime         time.Time
	lock             *flock.Flock
	dir              string
}

func (l *writer) Close() error {
	if err := l.lock.Unlock(); err != nil {
		_ = l.currentSegment.Close()

		return fmt.Errorf("error unlocking the log lock: %w", err)
	}

	if err := l.currentSegment.Close(); err != nil {
		return fmt.Errorf("closing Writer failed: %w", err)
	}

	return nil
}

func (l *writer) Write(entry []byte, options ...WriteOption) (time.Time, error) {
	t := l.now()

	if !t.After(l.lastTime) {
		t = l.lastTime.Add(time.Nanosecond)
	}

	if err := l.writeEntry(t, entry); err != nil {
		return time.Time{}, err
	}

	l.lastTime = t

	return t, nil
}

func (l *writer) writeEntry(t time.Time, entry []byte) error {
	timeBinary, err := t.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshaling entry time failed: %w", err)
	}

	if _, err = l.currentSegment.Write(timeBinary); err != nil {
		return fmt.Errorf("writing entry time failed: %w", err)
	}

	if err = binary.Write(l.currentSegment, binary.LittleEndian, uint32(len(entry))); err != nil {
		return fmt.Errorf("writing entry len failed: %w", err)
	}

	if _, err = l.currentSegment.Write(entry); err != nil {
		return fmt.Errorf("writing entry data failed: %w", err)
	}

	l.currentSizeBytes += int64(len(timeBinary)) + int64(len(entry)) + sizeOfUint32
	if l.currentSizeBytes > l.maxSizeBytes {
		if err := l.rollOver(t.Add(time.Nanosecond)); err != nil {
			return err
		}
	}

	return nil
}

func (l *writer) rollOver(start time.Time) error {
	if err := l.currentSegment.Close(); err != nil {
		return fmt.Errorf("error closing segment file: %w", err)
	}

	l.currentSizeBytes = 0

	f := path.Join(l.dir, segmentFilenameStartingAt(start))

	currentSegment, err := openFileForAppending(f)
	if err != nil {
		return err
	}

	l.currentSegment = currentSegment

	return nil
}
