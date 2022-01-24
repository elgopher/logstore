package main

import (
	"fmt"

	"github.com/elgopher/logstore/log"
)

// This example shows how to write an entry to a log.
func main() {
	l := log.New("/tmp/logstore")

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

	// Write appends an entry to the log, returning a unique monotonic Time
	t, err := writer.Write([]byte("entry"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Entry written with t=", t)
}
