-- +goose Up
-- Веха 3 (E3.1, E3.3), ядро: регистрация институции владельцем + модерация approve/reject.
-- Idempotency-Key, ETag/If-Match optimistic locking, тарифы, multipart-медиа — отложены
-- до отдельных задач (см. internal/catalog/usecase, internal/moderation).
CREATE TABLE catalog.institution_owners (
    institution_id UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    user_id        UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (institution_id, user_id)
);

CREATE INDEX institution_owners_user_id_idx ON catalog.institution_owners (user_id);

-- E3.1: журнал решений модерации — не общий лог действий, только модераторские мутации
-- (approve/reject и т.п.). actor_id NULL зарезервирован под системные действия (без человека-актора).
CREATE TABLE moderation.audit_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_type   TEXT NOT NULL CHECK (actor_type IN ('user', 'system')),
    actor_id     UUID,
    actor_role   TEXT,
    action       TEXT NOT NULL,
    target_type  TEXT NOT NULL,
    target_id    UUID NOT NULL,
    reason_code  TEXT,
    reason_text  TEXT,
    payload_diff JSONB,
    request_id   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX audit_log_target_idx ON moderation.audit_log (target_type, target_id);

-- +goose Down
DROP TABLE IF EXISTS moderation.audit_log;
DROP TABLE IF EXISTS catalog.institution_owners;
