package main

import (
	"context"
	"fmt"
	"io"
	log2 "log"
	"time"

	"github.com/jacekolszak/logstore/log"
	"github.com/jacekolszak/logstore/replicator"
)

func main() {
	main := openLog("/tmp/main")
	replica1 := openLog("/tmp/replica1")
	replica2 := openLog("/tmp/replica2")
	defer closeAll(main, replica1, replica2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go replicator.Start(ctx, main, []*log.Log{replica1, replica2}, replicator.SkipFailedReadAfter(time.Minute))

	for i := 0; i < 10; i++ {
		_, err := main.Append([]byte("e"))
		if err != nil {
			panic(err)
		}
	}

	reader, err := replicator.Reader([]*log.Log{main, replica1, replica2})
	if err != nil {
		panic(err)
	}

	for {
		t, out, err := reader.Read()
		if err != nil {
			panic(err)
		}
		fmt.Println(t, out)
	}
}

func openLog(dir string) *log.Log {
	l, err := log.Open(dir)
	if err != nil {
		panic(err)
	}
	return l
}

func closeAll(c ...io.Closer) {
	for _, closer := range c {
		if err := closer.Close(); err != nil {
			log2.Print(err)
		}
	}
}
