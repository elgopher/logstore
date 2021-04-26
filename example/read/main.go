package main

import (
	"errors"
	"fmt"

	"github.com/jacekolszak/logstore/log"
)

// This example reads all entries from log.
func main() {
	l := log.New("/tmp/logstore")

	reader, err := l.OpenReader()
	if err != nil {
		panic(err)
	}

	for {
		t, data, err := reader.Read()
		if errors.Is(err, log.ErrEOL) {
			if err = reader.Close(); err != nil {
				panic(err)
			}

			return
		}

		if err != nil {
			panic(err)
		}

		fmt.Printf("Entry read with t=%s,data=%s\n", t, data)
	}
}
