ALTER TABLE photo_submissions
    ADD COLUMN IF NOT EXISTS media_type VARCHAR(16) NOT NULL DEFAULT 'photo';

ALTER TABLE photo_submissions DROP CONSTRAINT IF EXISTS photo_submissions_media_type_check;
ALTER TABLE photo_submissions
    ADD CONSTRAINT photo_submissions_media_type_check CHECK (media_type IN ('photo', 'video'));

CREATE INDEX IF NOT EXISTS idx_photo_submissions_pending_created_at
    ON photo_submissions (created_at)
    WHERE status = 'pending';
