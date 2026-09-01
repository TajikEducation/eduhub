-- +goose Up
-- Схема moderation уже создана в 00001_bootstrap.sql (CREATE SCHEMA IF NOT EXISTS) — здесь
-- только первая таблица в ней.
CREATE TABLE moderation.audit_log (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id      UUID,
    actor_type    TEXT NOT NULL DEFAULT 'user' CHECK (actor_type IN ('user', 'system')),
    actor_role    TEXT,
    action        TEXT NOT NULL,
    target_type   TEXT NOT NULL,
    target_id     UUID NOT NULL,
    reason_code   TEXT,
    reason_text   TEXT,
    payload_diff  JSONB,
    request_id    TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_target_type_target_id_created_at
    ON moderation.audit_log (target_type, target_id, created_at DESC);
CREATE INDEX idx_audit_log_actor_id ON moderation.audit_log (actor_id);

-- +goose Down
DROP TABLE moderation.audit_log;
