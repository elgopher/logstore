// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import (
	"fmt"
	"os"
	"path"
)

func Open(dir string, options ...OpenOption) (*Log, error) {
	l := &Log{}

	for _, opt := range options {
		if opt == nil {
			continue
		}

		if err := opt(l); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	if err := mkdirIfMissing(dir); err != nil {
		return nil, err
	}

	return l, nil
}

func mkdirIfMissing(dir string) error {
	_, err := os.Stat(path.Join(dir))
	if os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0775); err != nil {
			return fmt.Errorf("cannot create directory: %w", err)
		}
	}

	return nil
}

type OpenOption func(*Log) error

type Log struct {
}

func (l *Log) Close() error {
	return nil
}
