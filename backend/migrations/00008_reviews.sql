-- +goose Up
-- Веха 4, ядро — сознательно упрощённая версия полного SRS-спека (docs/EduHub_Database_Schema.md,
-- reviews.reviews): одна общая оценка вместо 8 метрик с decay-агрегацией, без верификации через
-- auth.children (эта таблица ещё не построена — см. миграцию 00006), без dispute-workflow и
-- без transactional outbox/Redis Streams — синхронное обновление catalog.institutions при approve.
-- institution_id/user_id — по значению, без физического кросс-схемного FK (тот же принцип
-- владения схемами, что и для остальных cross-schema ссылок в докe).
CREATE TABLE reviews.reviews (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL,
    user_id        UUID NOT NULL,
    rating         SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    text           TEXT NOT NULL,
    reply          TEXT,
    replied_at     TIMESTAMPTZ,
    status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (institution_id, user_id)
);

CREATE INDEX reviews_institution_status_idx ON reviews.reviews (institution_id, status);

-- +goose Down
DROP TABLE reviews.reviews;
