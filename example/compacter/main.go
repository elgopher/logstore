package main

import (
	"context"
	"fmt"
	"time"

	"github.com/elgopher/logstore/compacter"
	"github.com/elgopher/logstore/log"
)

// This example shows how to compact log by removing old segments.
func main() {
	l := log.New("/tmp/logstore")

	reader, err := l.OpenReader()
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = reader.Close(); err != nil {
			panic(err)
		}
	}()

	// Old segments can be removed manually:
	oneHourAgo := time.Now().Add(-time.Hour)

	results, err := compacter.RemoveOldSegments(l, oneHourAgo)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Segments removed: %d", len(results.SegmentsRemoved))

	// Or they can be removed continuously by go-routine running in the background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := compacter.Start(ctx, l, compacter.Retention(time.Hour)); err != nil {
			panic(err)
		}
	}()

	const day = 24 * time.Hour

	time.Sleep(day)
}
