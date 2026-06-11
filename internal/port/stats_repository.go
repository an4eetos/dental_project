package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/admin"
)

// StatsRepository aggregates clinic usage metrics.
type StatsRepository interface {
	GetStatistics(ctx context.Context) (admin.Statistics, error)
}
