package providerquery

import "time"

// Clock is used in place of direct calls to [time.Now].
type Clock interface {
	Now() time.Time
}

// standardClock satisfies the interface Clock and uses [time.Now].
type standardClock struct{}

var _ Clock = (*standardClock)(nil)

// Now returns the current time.
func (c standardClock) Now() time.Time {
	return time.Now()
}
