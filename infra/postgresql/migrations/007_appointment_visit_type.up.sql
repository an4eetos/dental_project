ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS visit_type TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS doctor_notes TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS responded_by BIGINT REFERENCES users (id),
    ADD COLUMN IF NOT EXISTS doctor_reminder_1d_sent_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS doctor_reminder_1h_sent_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS zoom_missing_notified_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_appointments_confirmed_reminders_v2
    ON appointments (preferred_date, preferred_time)
    WHERE status = 'confirmed';

CREATE INDEX IF NOT EXISTS idx_appointments_video_missing_zoom
    ON appointments (preferred_date, preferred_time)
    WHERE status = 'confirmed'
      AND visit_type = 'video'
      AND (zoom_link IS NULL OR TRIM(zoom_link) = '');
