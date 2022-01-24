// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package codec

import (
	"fmt"
	"time"

	"github.com/elgopher/logstore/log"
)

type Format interface {
	// Encode can (optionally) append bytes to the passed output slice to reduce allocations.
	Encode(input interface{}, output []byte) (out []byte, err error)
	Decode(input []byte, output interface{}) error
}

func New(format Format) *Codec {
	if format == nil {
		panic("nil format")
	}

	return &Codec{format: format}
}

type Codec struct {
	format Format
}

func (c *Codec) Write(writer Writer, object interface{}) (time.Time, error) {
	if writer == nil {
		return time.Time{}, fmt.Errorf("nil writer: %w", log.ErrInvalidParameter)
	}

	data, err := c.encode(object)
	if err != nil {
		return time.Time{}, err
	}

	t, err := writer.Write(data)
	if err != nil {
		return time.Time{}, fmt.Errorf("write failed: %w", err)
	}

	return t, nil
}

type Writer interface {
	Write(entry []byte) (time.Time, error)
}

func (c *Codec) encode(object interface{}) ([]byte, error) {
	data, err := c.format.Encode(object, make([]byte, 0))
	if err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return data, nil
}

func (c *Codec) WriteWithTime(writer WriterWithTime, t time.Time, object interface{}) error {
	if writer == nil {
		return fmt.Errorf("nil writer: %w", log.ErrInvalidParameter)
	}

	data, err := c.encode(object)
	if err != nil {
		return err
	}

	if err := writer.WriteWithTime(t, data); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

type WriterWithTime interface {
	WriteWithTime(t time.Time, entry []byte) error
}

func (c *Codec) Read(reader Reader, object interface{}) (time.Time, error) {
	if reader == nil {
		return time.Time{}, fmt.Errorf("nil writer: %w", log.ErrInvalidParameter)
	}

	t, data, err := reader.Read()
	if err != nil {
		return time.Time{}, fmt.Errorf("read failed: %w", err)
	}

	if err = c.format.Decode(data, object); err != nil {
		return time.Time{}, fmt.Errorf("decoding failed: %w", err)
	}

	return t, nil
}

type Reader interface {
	Read() (time.Time, []byte, error)
}
