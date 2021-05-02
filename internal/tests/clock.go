// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package tests

import "time"

type Clock struct {
	CurrentTime *time.Time
}

func (c *Clock) MoveForwardOneHour() {
	c.MoveForward(time.Hour)
}

func (c *Clock) MoveForward(d time.Duration) {
	t := c.CurrentTime.Add(d)
	c.CurrentTime = &t
}

func (c *Clock) Now() time.Time {
	return *c.CurrentTime
}
