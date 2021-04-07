// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package codec

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/jacekolszak/logstore/log"
)

type Codec interface {
	// Encode will append to passed byte slice (and magnify it, if needed)
	Encode(input interface{}, output []byte) (out []byte, err error)
	Decode(input []byte, output interface{}) error
}

func New(c Codec, o ...Option) *Log {
	return &Log{
		codec:  c,
		buffer: make([]byte, 4096),
	}
}

type Option func()

type Log struct {
	codec  Codec
	mutex  sync.Mutex
	buffer []byte
}

func (l *Log) Append(s LogStore, v interface{}, opts ...log.AppendOption) (time.Time, error) {
	l.mutex.Lock() // lock is needed because buffer is reused. Faster will be a pool of buffers (sync.Pool)
	defer l.mutex.Unlock()

	bytes, err := l.codec.Encode(v, l.buffer)
	if err != nil {
		return time.Time{}, err
	}
	l.buffer = bytes // if buffer was magnified store it for future use

	t, err := s.Append(bytes, opts...)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func (l *Log) Read(r log.Reader, v interface{}, opts ...log.ReadOption) (time.Time, error) {
	for {
		t, data, err := r.Read()
		if err != nil {
			return time.Time{}, err
		}
		err = l.codec.Decode(data, v)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}
}

type LogStore interface {
	Append([]byte, ...log.AppendOption) (time.Time, error)
}

type JSON struct{}

func (JSON) Encode(input interface{}, output []byte) ([]byte, error) {
	// TODO should append to output instead of creating []byte each time (though JSON is CPU/memory inefficient so maybe this does not make sense)
	return json.Marshal(input)
}

func (JSON) Decode(input []byte, output interface{}) error {
	return json.Unmarshal(input, output)
}
