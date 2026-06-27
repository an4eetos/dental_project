CREATE TABLE IF NOT EXISTS content_items (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    access TEXT NOT NULL CHECK (access IN ('public', 'subscription')),
    published BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    blocks JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_items_published_sort ON content_items (published, sort_order);

CREATE TABLE IF NOT EXISTS content_media (
    id BIGSERIAL PRIMARY KEY,
    content_item_id BIGINT REFERENCES content_items(id) ON DELETE SET NULL,
    mime_type TEXT NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_media_item_id ON content_media(content_item_id);

INSERT INTO content_items (title, description, access, published, sort_order, blocks)
VALUES
    (
        'Балалардың ауыз қуысы гигиенасы',
        'Рекомендации стоматолога для будущих мам',
        'public',
        TRUE,
        1,
        '[{"type":"youtube","data":{"youtube_id":"zQZ3SGSwGBI"}}]'::jsonb
    ),
    (
        'Гигиена полости рта детей',
        'Рекомендации стоматолога для будущих мам',
        'public',
        TRUE,
        2,
        '[{"type":"youtube","data":{"youtube_id":"IFT7drSL35s"}}]'::jsonb
    ),
    (
        'Жүктілік кезіндегі ауыз қуысының гигиенасы',
        'Рекомендации стоматолога для будущих мам',
        'subscription',
        TRUE,
        3,
        '[{"type":"youtube","data":{"youtube_id":"FMU4zgGRbiE"}}]'::jsonb
    ),
    (
        'Гигиена полости рта беременных пациентов',
        'Рекомендации стоматолога для будущих мам',
        'subscription',
        TRUE,
        4,
        '[{"type":"youtube","data":{"youtube_id":"yKlH5tjZTxI"}}]'::jsonb
    );
