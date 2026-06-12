package appointment

import (
	"context"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/port"
)

// PurgeOutdated removes appointments whose scheduled time is in the past.
type PurgeOutdated struct {
	Appointments port.AppointmentRepository
	Clock        port.Clock
	Location     *time.Location
}

func (uc *PurgeOutdated) Execute(ctx context.Context) (int64, error) {
	now := uc.Clock.Now()
	loc := uc.Location
	if loc == nil {
		loc = time.UTC
	}
	return uc.Appointments.DeleteScheduledBefore(ctx, now, loc)
}
