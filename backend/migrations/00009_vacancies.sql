-- +goose Up
-- Веха 4, ядро — вакансии учреждений (FR-36, docs/EduHub_Database_Schema.md,
-- communications.vacancies). Сознательно не включает applicants/applications (отклики,
-- профиль соискателя) — вторая волна, не запрошена явно, см. CLAUDE.md NEVER-лист.
-- institution_id — по значению, без физического FK (тот же принцип, что reviews.reviews).
CREATE TABLE communications.vacancies (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL,
    title          JSONB NOT NULL,
    description    JSONB NOT NULL,
    requirements   JSONB,
    salary_from    INT,
    salary_to      INT,
    employment     JSONB NOT NULL,
    status         TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX vacancies_institution_status_idx ON communications.vacancies (institution_id, status);
CREATE INDEX vacancies_status_created_idx ON communications.vacancies (status, created_at DESC);

-- +goose Down
DROP TABLE communications.vacancies;
