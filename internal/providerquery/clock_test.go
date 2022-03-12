package providerquery

import "time"

// fixedClock returns a preset time.
type fixedClock struct{ now time.Time }

var _ Clock = (*fixedClock)(nil)

// Now returns a fixed time.
func (c fixedClock) Now() time.Time {
	return c.now
}
