// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package tests

import "time"

type Clock struct {
	CurrentTime *time.Time
}

func (c *Clock) MoveForward() {
	t := c.CurrentTime.Add(time.Hour)
	c.CurrentTime = &t
}

func (c *Clock) Now() time.Time {
	return *c.CurrentTime
}
