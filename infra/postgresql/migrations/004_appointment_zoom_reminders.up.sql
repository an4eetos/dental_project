ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS zoom_link TEXT,
    ADD COLUMN IF NOT EXISTS offer_sent_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reminder_1d_sent_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reminder_1h_sent_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_appointments_confirmed_reminders
    ON appointments (preferred_date, preferred_time)
    WHERE status = 'confirmed' AND zoom_link IS NOT NULL;
