package main

import (
	"fmt"

	"github.com/elgopher/logstore/codec"
	"github.com/elgopher/logstore/log"
)

// This example shows how to write and read events using high level API.
func main() {
	l := log.New("/tmp/logstore/json")

	// create codec using JSON format. All written entries will be encoded using JSON.
	jsonCodec := codec.New(codec.JSON())

	// open standard writer
	writer, err := l.OpenWriter()
	if err != nil {
		panic(err)
	}

	defer func() {
		err = writer.Close()
		if err != nil {
			panic(err)
		}
	}()

	for i := 0; i < 10; i++ {
		// create entry object which will be encoded to JSON during writing
		entry := Entry{
			SomeNumber: i,
			Author:     "Gopher",
		}
		// write entry object, instead of byte slice
		_, err = jsonCodec.Write(writer, entry)
		if err != nil {
			panic(fmt.Sprintf("write failed: %v", err))
		}
	}
}

type Entry struct {
	SomeNumber int
	Author     string
}
