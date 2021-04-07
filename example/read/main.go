package main

import (
	"fmt"

	"github.com/jacekolszak/logstore/log"
)

func main() {
	l, err := log.Open("/tmp/logstore")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = l.Close(); err != nil {
			panic(err)
		}
	}()

	reader := l.Reader() // read all entries until EOL (end of log - that is - last segment and EOF)
	for {
		t, data, err := reader.Read()
		if log.IsEOL(err) {
			return
		}
		if err != nil {
			// problem found. Can't go any further (all next entries are lost)
			// next time we will read next entries (probably, because more recent snapshot will be created).
			fmt.Println(err)
			return
		}
		fmt.Println("entry found: ", t, data)
	}

}
