package main

import (
	"fmt"

	"github.com/jacekolszak/logstore/codec"
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

	json := codec.New(codec.JSON{})

	eventToWrite := Event{
		Type: "SomethingHappened",
		ID:   "1234",
	}
	t, err := json.Append(l, eventToWrite)
	fmt.Println("json saved with t=", t)

	reader := l.Reader()
	eventToLoad := Event{}
	for {
		t, err := json.Read(reader, eventToLoad)
		if log.IsEOL(err) {
			return
		}
		if err != nil {
			panic(err)
		}
		fmt.Println("entry found: ", t, eventToLoad)
	}
}

type Event struct {
	Type string
	ID   string
}
