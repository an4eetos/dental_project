package clock

import "time"

// System implements port.Clock using the real system clock.
type System struct{}

func (System) Now() time.Time { return time.Now() }
