ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS preferred_visit_type TEXT NOT NULL DEFAULT '';
