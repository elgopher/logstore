// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package log

import "errors"

var (
	ErrEOL    = errors.New("eol (end of log):  reading log finished")
	ErrLocked = errors.New("log is already locked for writing")
)
