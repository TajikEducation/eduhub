-- +goose Up
CREATE TABLE platform.seed_refs (
    seed_ref     TEXT NOT NULL,
    entity_type  TEXT NOT NULL,
    entity_id    UUID NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (seed_ref, entity_type)
);

-- +goose Down
DROP TABLE platform.seed_refs;
