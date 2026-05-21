package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
)

// AppointmentRepository implements port.AppointmentRepository.
type AppointmentRepository struct {
	pool *pgxpool.Pool
}

func NewAppointmentRepository(pool *pgxpool.Pool) *AppointmentRepository {
	return &AppointmentRepository{pool: pool}
}

func (r *AppointmentRepository) Create(ctx context.Context, appt booking.Appointment) (booking.Appointment, error) {
	const q = `
INSERT INTO appointments (user_id, preferred_date, preferred_time, status, created_at)
VALUES ($1, $2::date, $3::time, $4, COALESCE($5, NOW()))
RETURNING id, user_id, preferred_date, preferred_time, status, created_at`

	row := r.pool.QueryRow(ctx, q,
		appt.UserID,
		appt.PreferredDate,
		formatTime(appt.PreferredTime),
		appt.Status.String(),
		appt.CreatedAt,
	)
	return scanAppointment(row)
}

func (r *AppointmentRepository) ListByUserID(ctx context.Context, userID int64) ([]booking.Appointment, error) {
	const q = `
SELECT id, user_id, preferred_date, preferred_time, status, created_at
FROM appointments
WHERE user_id = $1
ORDER BY preferred_date ASC, preferred_time ASC, created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []booking.Appointment
	for rows.Next() {
		appt, err := scanAppointment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, appt)
	}
	return out, rows.Err()
}

func (r *AppointmentRepository) ListAllWithPatients(ctx context.Context) ([]booking.AppointmentWithPatient, error) {
	const q = `
SELECT
    a.id, a.user_id, a.preferred_date, a.preferred_time, a.status, a.created_at,
    u.id, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, '')
FROM appointments a
JOIN users u ON u.id = a.user_id
ORDER BY a.preferred_date ASC, a.preferred_time ASC, a.created_at DESC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []booking.AppointmentWithPatient
	for rows.Next() {
		item, err := scanAppointmentWithPatient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanAppointmentWithPatient(row pgxRowScanner) (booking.AppointmentWithPatient, error) {
	var item booking.AppointmentWithPatient
	var status string
	var prefTime time.Time
	err := row.Scan(
		&item.Appointment.ID,
		&item.Appointment.UserID,
		&item.Appointment.PreferredDate,
		&prefTime,
		&status,
		&item.Appointment.CreatedAt,
		&item.Patient.ID,
		&item.Patient.TelegramID,
		&item.Patient.Username,
		&item.Patient.FirstName,
		&item.Patient.LastName,
	)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}
	item.Appointment.Status = booking.Status(status)
	item.Appointment.PreferredDate = truncateDateUTC(item.Appointment.PreferredDate)
	item.Appointment.PreferredTime = time.Date(0, 1, 1, prefTime.Hour(), prefTime.Minute(), prefTime.Second(), 0, time.UTC)
	return item, nil
}

func scanAppointment(row pgxRowScanner) (booking.Appointment, error) {
	var appt booking.Appointment
	var status string
	var prefTime time.Time
	err := row.Scan(
		&appt.ID,
		&appt.UserID,
		&appt.PreferredDate,
		&prefTime,
		&status,
		&appt.CreatedAt,
	)
	if err != nil {
		return booking.Appointment{}, err
	}
	appt.Status = booking.Status(status)
	appt.PreferredDate = truncateDateUTC(appt.PreferredDate)
	appt.PreferredTime = time.Date(0, 1, 1, prefTime.Hour(), prefTime.Minute(), prefTime.Second(), 0, time.UTC)
	return appt, nil
}

func truncateDateUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

type pgxRowScanner interface {
	Scan(dest ...any) error
}

func formatTime(t time.Time) string {
	return t.Format("15:04:05")
}
