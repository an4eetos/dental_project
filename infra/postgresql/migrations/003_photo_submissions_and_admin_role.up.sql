ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('patient', 'doctor', 'admin'));

CREATE TABLE IF NOT EXISTS photo_submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    telegram_file_id TEXT NOT NULL,
    image_data BYTEA NOT NULL,
    image_mime TEXT NOT NULL DEFAULT 'image/jpeg',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    ai_draft JSONB,
    doctor_response TEXT,
    responded_by BIGINT REFERENCES users(id),
    responded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT photo_submissions_status_check CHECK (status IN ('pending', 'answered'))
);

CREATE INDEX IF NOT EXISTS idx_photo_submissions_user_id ON photo_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_photo_submissions_status ON photo_submissions(status);
CREATE INDEX IF NOT EXISTS idx_photo_submissions_created_at ON photo_submissions(created_at DESC);
