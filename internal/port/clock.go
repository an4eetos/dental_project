package port

import "time"

// Clock abstracts time for testable usecases.
type Clock interface {
	Now() time.Time
}
