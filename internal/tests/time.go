// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package tests

import (
	"time"
)

func MustTime(rfc3339 string) time.Time {
	tt, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		panic(err)
	}

	return tt
}
