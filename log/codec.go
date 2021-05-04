package log

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

func decodeEntry(reader io.Reader) (time.Time, []byte, error) {
	t := time.Time{}

	bytes := make([]byte, 15)

	_, err := io.ReadAtLeast(reader, bytes, 15)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("reading entry time failed: %w", err)
	}

	if err = t.UnmarshalBinary(bytes[:15]); err != nil {
		return time.Time{}, nil, fmt.Errorf("unmarshaling entry time failed: %w", err)
	}

	var length uint32
	if err = binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return time.Time{}, nil, fmt.Errorf("reading entry len failed: %w", err)
	}

	data := make([]byte, length)
	if _, err = reader.Read(data); err != nil {
		return time.Time{}, nil, fmt.Errorf("reading entry data failed: %w", err)
	}

	return t, data, nil
}

func encodeEntry(writer io.Writer, t time.Time, entry []byte) error {
	timeBinary, err := t.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshaling entry time failed: %w", err)
	}

	if _, err = writer.Write(timeBinary); err != nil {
		return fmt.Errorf("writing entry time failed: %w", err)
	}

	if err = binary.Write(writer, binary.LittleEndian, uint32(len(entry))); err != nil {
		return fmt.Errorf("writing entry len failed: %w", err)
	}

	if _, err = writer.Write(entry); err != nil {
		return fmt.Errorf("writing entry data failed: %w", err)
	}

	return nil
}
