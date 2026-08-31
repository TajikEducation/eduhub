-- +goose Up
-- Веха 5, ядро — упрощённая версия SRS-доски вакансий для соискателей
-- (docs/EduHub_Database_Schema.md, communications.applicants/applicant_achievements/applications):
-- без employer_responses (приглашения работодателя) — не запрошено явно, тот же принцип
-- "core slice", что у reviews/vacancies/chat в этой сессии. hide_contacts/visibility —
-- как в SRS; фильтрация контактов делается в SQL-проекции репозитория, не в DTO-маппинге
-- после чтения (см. комментарий в docs, E5.5).
CREATE TABLE communications.applicants (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL UNIQUE,
    name           JSONB NOT NULL,
    photo_url      TEXT,
    position       JSONB NOT NULL,
    bio            JSONB,
    education      JSONB,
    experience     JSONB,
    skills         JSONB,
    email          TEXT,
    phone          TEXT,
    cv_s3_key      TEXT,
    visibility     TEXT NOT NULL DEFAULT 'draft' CHECK (visibility IN ('draft', 'on_response', 'public')),
    hide_contacts  BOOL NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX applicants_visibility_public_idx ON communications.applicants (visibility) WHERE visibility = 'public';

CREATE TABLE communications.applicant_achievements (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_id  UUID NOT NULL REFERENCES communications.applicants(id) ON DELETE CASCADE,
    title         TEXT NOT NULL,
    year          INT,
    category      TEXT CHECK (category IN ('gold', 'silver', 'bronze', 'special')),
    description   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX applicant_achievements_applicant_idx ON communications.applicant_achievements (applicant_id);

CREATE TABLE communications.applications (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_id  UUID NOT NULL REFERENCES communications.applicants(id) ON DELETE CASCADE,
    vacancy_id    UUID NOT NULL REFERENCES communications.vacancies(id) ON DELETE CASCADE,
    status        TEXT NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'viewed', 'closed')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (applicant_id, vacancy_id)
);

-- +goose Down
DROP TABLE communications.applications;
DROP TABLE communications.applicant_achievements;
DROP TABLE communications.applicants;
