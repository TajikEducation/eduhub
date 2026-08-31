-- +goose Up
-- Веха 5, ядро — упрощённая версия SRS-чата (docs/EduHub_Database_Schema.md,
-- communications.conversations/messages): REST-опрос (поллинг на фронте), без WebSocket/
-- Redis Pub/Sub (план E5.2) и без push-уведомлений — те же осознанные упрощения, что у
-- reviews/vacancies в этой веху. participant_*_id — по значению, без физического FK
-- (участник может быть user ИЛИ institution — полиморфный слот).
CREATE TABLE communications.conversations (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_a_type          TEXT NOT NULL CHECK (participant_a_type IN ('user', 'institution')),
    participant_a_id            UUID NOT NULL,
    participant_b_type          TEXT NOT NULL CHECK (participant_b_type IN ('user', 'institution')),
    participant_b_id            UUID NOT NULL,
    participant_a_last_read_at  TIMESTAMPTZ,
    participant_b_last_read_at  TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (participant_a_type, participant_a_id, participant_b_type, participant_b_id)
);

CREATE TABLE communications.messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES communications.conversations(id) ON DELETE CASCADE,
    sender_type     TEXT NOT NULL CHECK (sender_type IN ('user', 'institution')),
    sender_id       UUID NOT NULL,
    body            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX messages_conversation_created_idx ON communications.messages (conversation_id, created_at);

-- +goose Down
DROP TABLE communications.messages;
DROP TABLE communications.conversations;
