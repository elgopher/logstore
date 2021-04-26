// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

func (l *Log) openReader(options []OpenReaderOption) (Reader, error) {
	settings := &ReaderSettings{}

	for _, applyOption := range options {
		if applyOption == nil {
			continue
		}

		if err := applyOption(settings); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	f, err := os.Open(path.Join(l.dir, "segment.data"))
	if os.IsNotExist(err) {
		return &emptyLogReader{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("opening segment file failed: %w", err)
	}

	return &singleSegmentReader{
		currentSegment: f,
	}, nil
}

type emptyLogReader struct{}

func (r *emptyLogReader) Read() (time.Time, []byte, error) {
	return time.Time{}, nil, ErrEOL
}

func (r *emptyLogReader) Close() error {
	return nil
}

type singleSegmentReader struct {
	currentSegment *os.File
}

func (r *singleSegmentReader) Read() (time.Time, []byte, error) {
	t := time.Time{}
	bytes := make([]byte, 15)

	_, err := io.ReadAtLeast(r.currentSegment, bytes, 15)
	if err != nil {
		return time.Time{}, nil, wrapReadError("reading entry time failed", err)
	}

	if err = t.UnmarshalBinary(bytes[:15]); err != nil {
		return time.Time{}, nil, wrapReadError("unmarshaling entry time failed", err)
	}

	var length uint32
	if err = binary.Read(r.currentSegment, binary.LittleEndian, &length); err != nil {
		return time.Time{}, nil, wrapReadError("reading entry len failed", err)
	}

	data := make([]byte, length)
	if _, err = r.currentSegment.Read(data); err != nil {
		return time.Time{}, nil, wrapReadError("reading entry data failed", err)
	}

	return t, data, nil
}

func wrapReadError(msg string, err error) error {
	if errors.Is(err, io.EOF) {
		return ErrEOL
	}

	return fmt.Errorf("%s: %w", msg, err)
}

func (r *singleSegmentReader) Close() error {
	if err := r.currentSegment.Close(); err != nil {
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
