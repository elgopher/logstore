// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package replicator

import (
	"context"
	"time"

	"github.com/jacekolszak/logstore/log"
)

func Start(ctx context.Context, from *log.Log, to []*log.Log, options ...Option) {
	var data [4096]byte
	reader := from.Reader(log.DontReportEol)
	for {
		t, data, err := reader.Read(log.AppendToBuffer(data[:0]))
		if err != nil {
			// TODO LOG ERROR, BUT CONTINUE UNTIL READ RETURNS SUCCCESS, or SkipFailedReadAfter was used and timeout
			time.Sleep(1 * time.Second)
		}
		_, err = to[0].Append(data, log.ForceTime(t))
		if err != nil {
			// TODO LOG ERROR, WRITE WITH ERROR SHOULD OPEN NEW SEGMENT ON NEXT WRITE
		}
	}
}

type Option func()

func SkipFailedReadAfter(timeout time.Duration) Option {
	return func() {

	}
}

// Implementation is very simple. For each log reader is created. Read on returned Reader reads from all logs
// and returns the oldest and without errors.
//
// More complicated cases
// is in r1, but corrupted, missing in r2 -> error
// is in r1 and in r2 -> take from r1
// is in r1, but corrupted, is in r2 -> take from r2
func Reader(replicatedLogs []*log.Log, options ...log.ReaderOption) (log.Reader, error) {
	return &reader{
		logs:    replicatedLogs,
		options: options,
	}, nil
}

type reader struct {
	logs    []*log.Log
	options []log.ReaderOption
}

func (r *reader) Read(...log.ReadOption) (time.Time, []byte, error) {
	// error returned from reader?
	panic("implement me")
}

func (r *reader) Close() error {
	panic("implement me")
}
