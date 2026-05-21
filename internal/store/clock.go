package store

import "time"

type Clock interface {
	Now() time.Time
}

type ClockProvider struct{}

func (c ClockProvider) Now() time.Time {
	return time.Now()
}
