package main

import (
	"time"

	"github.com/jacekolszak/logstore/compacter"
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

	oneHourAgo := time.Now().Add(-time.Hour)
	_, err = compacter.RemoveOldSegments(l, oneHourAgo)
	if err != nil {
		panic(err)
	}

}
