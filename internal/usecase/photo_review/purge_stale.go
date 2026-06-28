package photoreview

import (
	"context"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/port"
)

// PurgeStale removes pending submissions older than the configured max age.
type PurgeStale struct {
	Submissions port.PhotoSubmissionRepository
	Clock       port.Clock
	MaxAge      time.Duration
}

func (uc *PurgeStale) Execute(ctx context.Context) (int64, error) {
	maxAge := uc.MaxAge
	if maxAge <= 0 {
		maxAge = 10 * 24 * time.Hour
	}
	cutoff := uc.Clock.Now().Add(-maxAge)
	return uc.Submissions.DeletePendingOlderThan(ctx, cutoff)
}
