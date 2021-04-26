package main

import (
	"fmt"

	"github.com/jacekolszak/logstore/log"
)

// This example shows how to write an entry to a log.
func main() {
	l := log.New("/tmp/logstore")

	writer, err := l.OpenWriter()
	if err != nil {
		panic(err)
	}

	t, err := writer.Write([]byte("entry"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Entry written with t=", t)

	err = writer.Close()
	if err != nil {
		panic(err)
	}
}
