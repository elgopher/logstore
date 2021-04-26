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
	settings := &WriterSettings{
		now: time.Now,
	}

	for _, applyOption := range options {
		if applyOption == nil {
			continue
		}

		if err := applyOption(settings); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	if err := mkdirIfMissing(l.dir); err != nil {
		return nil, err
	}

	lockFile := path.Join(l.dir, "log.lock")
	lock := flock.New(lockFile)

	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("error trying to lock log for writing: %w", err)
	}

	if !locked {
		return nil, ErrLocked
	}

	lastTime, err := l.readLastTime()
	if err != nil {
		return nil, err
	}

	file := path.Join(l.dir, "segment.data")

	currentSegment, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664)
	if err != nil {
		return nil, fmt.Errorf("error opening segment file %s for write: %w", file, err)
	}

	return &writer{
		now:            settings.now,
		lastTime:       lastTime,
		currentSegment: currentSegment,
		lock:           lock,
	}, nil
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

type writer struct {
	now            func() time.Time
	currentSegment *os.File
	lastTime       time.Time
	lock           *flock.Flock
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

	return nil
}
