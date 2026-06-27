package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

const appointmentSelectCols = `
    a.id, a.user_id, a.preferred_date, a.preferred_time, a.status,
    COALESCE(a.visit_type, ''), COALESCE(a.zoom_link, ''), COALESCE(a.doctor_notes, ''),
    a.responded_by, a.offer_sent_at,
    a.reminder_1d_sent_at, a.reminder_1h_sent_at,
    a.doctor_reminder_1d_sent_at, a.doctor_reminder_1h_sent_at,
    a.zoom_missing_notified_at, a.created_at`

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
RETURNING id, user_id, preferred_date, preferred_time, status,
    COALESCE(visit_type, ''), COALESCE(zoom_link, ''), COALESCE(doctor_notes, ''),
    responded_by, offer_sent_at,
    reminder_1d_sent_at, reminder_1h_sent_at,
    doctor_reminder_1d_sent_at, doctor_reminder_1h_sent_at,
    zoom_missing_notified_at, created_at`

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
SELECT id, user_id, preferred_date, preferred_time, status,
    COALESCE(visit_type, ''), COALESCE(zoom_link, ''), COALESCE(doctor_notes, ''),
    responded_by, offer_sent_at,
    reminder_1d_sent_at, reminder_1h_sent_at,
    doctor_reminder_1d_sent_at, doctor_reminder_1h_sent_at,
    zoom_missing_notified_at, created_at
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
SELECT` + appointmentSelectCols + `,
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

func (r *AppointmentRepository) GetWithPatientByID(ctx context.Context, id int64) (booking.AppointmentWithPatient, error) {
	const q = `
SELECT` + appointmentSelectCols + `,
    u.id, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, '')
FROM appointments a
JOIN users u ON u.id = a.user_id
WHERE a.id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	item, err := scanAppointmentWithPatient(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return booking.AppointmentWithPatient{}, domainerrors.ErrAppointmentNotFound
	}
	return item, err
}

func (r *AppointmentRepository) UpdateRespond(ctx context.Context, update booking.RespondUpdate) error {
	status := booking.StatusConfirmed.String()
	if update.Decision == booking.DecisionReject {
		status = booking.StatusRejected.String()
	}

	const q = `
UPDATE appointments
SET preferred_date = $2::date,
    preferred_time = $3::time,
    status = $4,
    visit_type = $5,
    zoom_link = $6,
    doctor_notes = $7,
    responded_by = $8,
    offer_sent_at = $9,
    reminder_1d_sent_at = NULL,
    reminder_1h_sent_at = NULL,
    doctor_reminder_1d_sent_at = NULL,
    doctor_reminder_1h_sent_at = NULL,
    zoom_missing_notified_at = NULL
WHERE id = $1 AND status = 'pending'`

	tag, err := r.pool.Exec(ctx, q,
		update.ID,
		update.PreferredDate,
		formatTime(update.PreferredTime),
		status,
		update.VisitType.String(),
		update.ZoomLink,
		update.DoctorNotes,
		update.RespondedBy,
		update.RespondedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.ErrAppointmentNotPending
	}
	return nil
}

func (r *AppointmentRepository) UpdateZoomLink(ctx context.Context, update booking.ZoomLinkUpdate) error {
	const q = `
UPDATE appointments
SET zoom_link = $2,
    zoom_missing_notified_at = NULL,
    reminder_1d_sent_at = NULL,
    reminder_1h_sent_at = NULL,
    doctor_reminder_1d_sent_at = NULL,
    doctor_reminder_1h_sent_at = NULL
WHERE id = $1
  AND status = 'confirmed'
  AND visit_type = 'video'`

	tag, err := r.pool.Exec(ctx, q, update.ID, update.ZoomLink)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.ErrAppointmentNotVideo
	}
	return nil
}

func (r *AppointmentRepository) ListConfirmedForReminders(ctx context.Context) ([]booking.AppointmentWithPatient, error) {
	const q = `
SELECT` + appointmentSelectCols + `,
    u.id, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, '')
FROM appointments a
JOIN users u ON u.id = a.user_id
WHERE a.status = 'confirmed'
ORDER BY a.preferred_date ASC, a.preferred_time ASC`

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

func (r *AppointmentRepository) ListVideoMissingZoom(ctx context.Context) ([]booking.AppointmentWithPatient, error) {
	const q = `
SELECT` + appointmentSelectCols + `,
    u.id, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, '')
FROM appointments a
JOIN users u ON u.id = a.user_id
WHERE a.status = 'confirmed'
  AND a.visit_type = 'video'
  AND (a.zoom_link IS NULL OR TRIM(a.zoom_link) = '')
  AND a.zoom_missing_notified_at IS NULL
ORDER BY a.preferred_date ASC, a.preferred_time ASC`

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

func (r *AppointmentRepository) DeleteScheduledBefore(
	ctx context.Context,
	before time.Time,
	loc *time.Location,
) (int64, error) {
	if loc == nil {
		loc = time.UTC
	}
	local := before.In(loc)
	y, m, d := local.Date()
	cutoffDate := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	const q = `
DELETE FROM appointments
WHERE preferred_date < $1::date
   OR (preferred_date = $1::date AND preferred_time < $2::time)`

	tag, err := r.pool.Exec(ctx, q, cutoffDate, formatTime(local))
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *AppointmentRepository) MarkReminderSent(
	ctx context.Context,
	id int64,
	kind booking.ReminderKind,
	at time.Time,
) error {
	var q string
	switch kind {
	case booking.ReminderOneDay:
		q = `UPDATE appointments SET reminder_1d_sent_at = $2 WHERE id = $1 AND reminder_1d_sent_at IS NULL`
	case booking.ReminderOneHour:
		q = `UPDATE appointments SET reminder_1h_sent_at = $2 WHERE id = $1 AND reminder_1h_sent_at IS NULL`
	case booking.DoctorReminderOneDay:
		q = `UPDATE appointments SET doctor_reminder_1d_sent_at = $2 WHERE id = $1 AND doctor_reminder_1d_sent_at IS NULL`
	case booking.DoctorReminderOneHour:
		q = `UPDATE appointments SET doctor_reminder_1h_sent_at = $2 WHERE id = $1 AND doctor_reminder_1h_sent_at IS NULL`
	case booking.ZoomMissingNotified:
		q = `UPDATE appointments SET zoom_missing_notified_at = $2 WHERE id = $1 AND zoom_missing_notified_at IS NULL`
	}
	if q == "" {
		return nil
	}

	tag, err := r.pool.Exec(ctx, q, id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	return nil
}

func scanAppointmentWithPatient(row pgxRowScanner) (booking.AppointmentWithPatient, error) {
	var item booking.AppointmentWithPatient
	var status string
	var visitType string
	var prefTime time.Time
	err := row.Scan(
		&item.Appointment.ID,
		&item.Appointment.UserID,
		&item.Appointment.PreferredDate,
		&prefTime,
		&status,
		&visitType,
		&item.Appointment.ZoomLink,
		&item.Appointment.DoctorNotes,
		&item.Appointment.RespondedBy,
		&item.Appointment.OfferSentAt,
		&item.Appointment.Reminder1dAt,
		&item.Appointment.Reminder1hAt,
		&item.Appointment.DoctorReminder1dAt,
		&item.Appointment.DoctorReminder1hAt,
		&item.Appointment.ZoomMissingNotifiedAt,
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
	item.Appointment.VisitType = booking.VisitType(visitType)
	item.Appointment.PreferredDate = truncateDateUTC(item.Appointment.PreferredDate)
	item.Appointment.PreferredTime = time.Date(0, 1, 1, prefTime.Hour(), prefTime.Minute(), prefTime.Second(), 0, time.UTC)
	return item, nil
}

func scanAppointment(row pgxRowScanner) (booking.Appointment, error) {
	var appt booking.Appointment
	var status string
	var visitType string
	var prefTime time.Time
	err := row.Scan(
		&appt.ID,
		&appt.UserID,
		&appt.PreferredDate,
		&prefTime,
		&status,
		&visitType,
		&appt.ZoomLink,
		&appt.DoctorNotes,
		&appt.RespondedBy,
		&appt.OfferSentAt,
		&appt.Reminder1dAt,
		&appt.Reminder1hAt,
		&appt.DoctorReminder1dAt,
		&appt.DoctorReminder1hAt,
		&appt.ZoomMissingNotifiedAt,
		&appt.CreatedAt,
	)
	if err != nil {
		return booking.Appointment{}, err
	}
	appt.Status = booking.Status(status)
	appt.VisitType = booking.VisitType(visitType)
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
