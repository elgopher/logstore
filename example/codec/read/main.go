package main

import (
	"errors"
	"fmt"

	"github.com/jacekolszak/logstore/codec"
	"github.com/jacekolszak/logstore/log"
)

// This example shows how to write and read events using high level API.
func main() {
	l := log.New("/tmp/logstore/json")

	// create codec using JSON format. All read entries will be decoded using JSON.
	jsonCodec := codec.New(codec.JSON())

	// open standard reader
	reader, err := l.OpenReader()
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = reader.Close(); err != nil {
			panic(err)
		}
	}()

	// create struct instance which will be used as a result for decoding
	entry := Entry{}

	for {
		// read object, entry is populated with decoded data
		t, err := jsonCodec.Read(reader, &entry)
		if errors.Is(err, log.ErrEOL) {
			return
		}

		if err != nil {
			panic(fmt.Sprintf("read failed: %v", err))
		}

		fmt.Printf("Entry read %s %+v\n", t, entry)
	}
}

type Entry struct {
	SomeNumber int
	Author     string
}
