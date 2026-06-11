package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/anuarkuanysh/dental_project/internal/domain/admin"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
)

// StatsRepository implements port.StatsRepository.
type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

func (r *StatsRepository) GetStatistics(ctx context.Context) (admin.Statistics, error) {
	const q = `
SELECT
    (SELECT COUNT(*) FROM users),
    (SELECT COUNT(*) FROM users WHERE role = $1),
    (SELECT COUNT(*) FROM users WHERE role = $2),
    (SELECT COUNT(*) FROM users WHERE role = $3),
    (SELECT COUNT(*) FROM photo_submissions),
    (SELECT COUNT(*) FROM photo_submissions WHERE status = $4),
    (SELECT COUNT(*) FROM photo_submissions WHERE status = $5),
    (SELECT COUNT(*) FROM appointments),
    (SELECT COUNT(*) FROM appointments WHERE status = $6),
    (SELECT COUNT(*) FROM appointments WHERE status = $7),
    (SELECT COUNT(*) FROM appointments WHERE status = $8)`

	var stats admin.Statistics
	err := r.pool.QueryRow(ctx, q,
		identity.RolePatient.String(),
		identity.RoleDoctor.String(),
		identity.RoleAdmin.String(),
		photoreview.StatusPending.String(),
		photoreview.StatusAnswered.String(),
		booking.StatusPending.String(),
		booking.StatusConfirmed.String(),
		booking.StatusCancelled.String(),
	).Scan(
		&stats.TotalUsers,
		&stats.TotalPatients,
		&stats.TotalDoctors,
		&stats.TotalAdmins,
		&stats.TotalPhotoSubmissions,
		&stats.PendingPhotoSubmissions,
		&stats.AnsweredPhotoSubmissions,
		&stats.TotalAppointments,
		&stats.PendingAppointments,
		&stats.ConfirmedAppointments,
		&stats.CancelledAppointments,
	)
	return stats, err
}
