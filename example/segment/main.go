package main

import (
	"fmt"

	"github.com/elgopher/logstore/log"
)

// This example reads information about log segments.
func main() {
	l := log.New("/tmp/logstore")

	segments, err := l.Segments()
	if err != nil {
		panic(err)
	}

	for _, segment := range segments {
		fmt.Printf("Segment %+v\n", segment)
	}
}
