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

	t, err := l.Append([]byte("entry"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Entry written with t=", t)

	err = l.Close()
	if err != nil {
		panic(err)
	}
}
