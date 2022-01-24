// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package codec_test

import (
	"errors"
	"testing"
	"time"

	"github.com/elgopher/logstore/codec"
	"github.com/elgopher/logstore/internal/tests"
	"github.com/elgopher/logstore/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should panic when format is nil", func(t *testing.T) {
		assert.Panics(t, func() {
			codec.New(nil)
		})
	})
}

func TestCodec_Read(t *testing.T) {
	t.Run("should return error when trying to read from nil reader", func(t *testing.T) {
		c := codec.New(&messageFormat{})
		_, err := c.Read(nil, "data")
		assert.ErrorIs(t, err, log.ErrInvalidParameter)
	})

	t.Run("should return error when there is no more to read", func(t *testing.T) {
		reader := tests.OpenLogReader(t)
		c := codec.New(&messageFormat{})
		var out *message
		// when
		_, err := c.Read(reader, out)
		// then
		assert.ErrorIs(t, err, log.ErrEOL)
	})

	t.Run("should return error when object cannot be unmarshalled", func(t *testing.T) {
		l, writer := tests.OpenLogWithWriter(t)
		tests.WriteEntry(t, writer, 4)
		_ = writer.Close()

		reader := tests.OpenReader(t, l)
		c := codec.New(&messageFormat{})
		var notMessage *int
		// when
		_, err := c.Read(reader, notMessage)
		// then
		assert.ErrorIs(t, err, errOutputIsNotMessage)
	})
}

type writer interface {
	Write(entry []byte) (time.Time, error)
	WriteWithTime(t time.Time, entry []byte) error
}

func TestCodec_Write(t *testing.T) {
	functions := map[string]func(c *codec.Codec, writer writer, object interface{}) (time.Time, error){
		"Write": func(c *codec.Codec, writer writer, object interface{}) (time.Time, error) {
			return c.Write(writer, object) // nolint:wrapcheck
		},
		"WriteWithTime": func(c *codec.Codec, writer writer, object interface{}) (time.Time, error) {
			now := time.Now()

			return now, c.WriteWithTime(writer, now, object) // nolint:wrapcheck
		},
	}

	for name, write := range functions {
		t.Run(name, func(t *testing.T) {
			t.Run("should return error when trying to write to nil writer", func(t *testing.T) {
				c := codec.New(&messageFormat{})
				_, err := write(c, nil, "data")
				assert.ErrorIs(t, err, log.ErrInvalidParameter)
			})

			t.Run("should return error when object cannot be marshalled", func(t *testing.T) {
				writer := tests.OpenLogWriter(t)
				c := codec.New(&messageFormat{})
				notMessage := 1
				_, err := write(c, writer, notMessage)
				// when
				assert.ErrorIs(t, err, errInputIsNotMessage)
			})

			t.Run("should write entry", func(t *testing.T) {
				l, writer := tests.OpenLogWithWriter(t)
				c := codec.New(&messageFormat{})
				msg := message{text: "data"}
				// when
				writeTime, err := write(c, writer, msg)
				require.NoError(t, err)
				// then
				reader := tests.OpenReader(t, l)

				var output message
				readTime, err := c.Read(reader, &output)
				require.NoError(t, err)

				assert.True(t, writeTime.Equal(readTime))
				require.NotNil(t, output)
				assert.Equal(t, msg, output)
			})
		})
	}
}

var errInputIsNotMessage = errors.New("input is not a message")
var errOutputIsNotMessage = errors.New("output is not a message")

type message struct {
	text string
}

type messageFormat struct {
}

func (f *messageFormat) Encode(input interface{}, output []byte) (out []byte, err error) {
	s, ok := input.(message)
	if !ok {
		return nil, errInputIsNotMessage
	}

	return []byte(s.text), nil
}

func (f *messageFormat) Decode(input []byte, output interface{}) error {
	msg, ok := output.(*message)
	if !ok {
		return errOutputIsNotMessage
	}

	msg.text = string(input)

	return nil
}
