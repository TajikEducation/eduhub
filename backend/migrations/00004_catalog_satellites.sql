-- +goose Up
CREATE TABLE catalog.institution_staff (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id  UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    name            JSONB NOT NULL,
    role_type       TEXT NOT NULL,
    role_label      JSONB NOT NULL,
    subject         JSONB,
    photo_url       TEXT,
    exp             TEXT,
    bio             JSONB,
    education       JSONB,
    email           TEXT,
    phone           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_institution_staff_institution_id
    ON catalog.institution_staff (institution_id);

CREATE TABLE catalog.achievements (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type   TEXT NOT NULL CHECK (owner_type IN ('institution', 'staff', 'student')),
    owner_id     UUID NOT NULL,
    title        JSONB NOT NULL,
    year         INT NOT NULL,
    category     TEXT NOT NULL,
    description  JSONB NOT NULL,
    links        JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_achievements_owner
    ON catalog.achievements (owner_type, owner_id);

CREATE TABLE catalog.institution_gallery (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id  UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    s3_key          TEXT NOT NULL,
    label           JSONB,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_institution_gallery_institution_id_sort_order
    ON catalog.institution_gallery (institution_id, sort_order);

CREATE TABLE catalog.institution_transport_routes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id  UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    type            TEXT NOT NULL CHECK (type IN ('own_bus', 'minibus', 'taxi', 'parent_coop', 'other')),
    label           JSONB,
    areas           JSONB,
    cost            INT CHECK (cost IS NULL OR cost >= 0),
    cost_period     TEXT NOT NULL DEFAULT 'month' CHECK (cost_period IN ('month', 'day', 'trip')),
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_institution_transport_routes_institution_id_sort_order
    ON catalog.institution_transport_routes (institution_id, sort_order);

CREATE TABLE catalog.institution_meal_plans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id  UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    meal_type       TEXT NOT NULL CHECK (meal_type IN ('hot', 'breakfast', 'buffet', 'other')),
    label           JSONB,
    cost            INT CHECK (cost IS NULL OR cost >= 0),
    cost_period     TEXT NOT NULL DEFAULT 'month' CHECK (cost_period IN ('month', 'day')),
    halal           BOOL,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_institution_meal_plans_institution_id_sort_order
    ON catalog.institution_meal_plans (institution_id, sort_order);

CREATE TABLE catalog.institution_alumni (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id  UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    name            JSONB NOT NULL,
    photo_url       TEXT,
    grad_year       INT NOT NULL,
    now_label       JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_institution_alumni_institution_id
    ON catalog.institution_alumni (institution_id);

CREATE TABLE catalog.news_articles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id  UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    title           JSONB NOT NULL,
    category        JSONB,
    cover_s3_key    TEXT,
    video_url       TEXT,
    content         JSONB NOT NULL,
    tags            JSONB,
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('published', 'draft')),
    views_count     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_news_articles_institution_id_status_created_at
    ON catalog.news_articles (institution_id, status, created_at DESC);

-- Пустая на этой вехе, заполняется вехой 4 через RatingSync.
CREATE TABLE catalog.institution_metrics (
    institution_id  UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    metric_key      TEXT NOT NULL,
    weighted_avg    NUMERIC(3, 2),
    review_count    INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (institution_id, metric_key)
);

CREATE TABLE catalog.institution_owner_verifications (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id       UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE CASCADE,
    submitted_by         UUID NOT NULL,
    document_type        TEXT NOT NULL CHECK (document_type IN ('license', 'state_status_confirmation', 'appointment_proof', 'business_registration', 'manual_exception', 'other')),
    document_s3_key      TEXT,
    license_no_claimed   TEXT,
    verification_notes   TEXT,
    status               TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by          UUID,
    reviewed_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_institution_owner_verifications_institution_id_created_at
    ON catalog.institution_owner_verifications (institution_id, created_at DESC);

CREATE INDEX idx_institution_owner_verifications_status
    ON catalog.institution_owner_verifications (status);

-- +goose Down
DROP TABLE catalog.institution_owner_verifications;
DROP TABLE catalog.institution_metrics;
DROP TABLE catalog.news_articles;
DROP TABLE catalog.institution_alumni;
DROP TABLE catalog.institution_meal_plans;
DROP TABLE catalog.institution_transport_routes;
DROP TABLE catalog.institution_gallery;
DROP TABLE catalog.achievements;
DROP TABLE catalog.institution_staff;
