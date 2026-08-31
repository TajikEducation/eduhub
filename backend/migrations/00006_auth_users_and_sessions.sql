-- +goose Up
CREATE TABLE auth.users (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                  TEXT NOT NULL,
    display_name           TEXT,
    locale                 TEXT NOT NULL DEFAULT 'ru',
    phone                  TEXT,
    password_hash          TEXT,
    role                   TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('guest', 'user', 'institution', 'moderator', 'admin')),
    status                 TEXT NOT NULL DEFAULT 'unverified' CHECK (status IN ('unverified', 'active', 'banned', 'deleted')),
    email_verified_at      TIMESTAMPTZ,
    consent_at             TIMESTAMPTZ NOT NULL,
    consent_version        TEXT NOT NULL,
    failed_login_count     INT NOT NULL DEFAULT 0,
    locked_until           TIMESTAMPTZ,
    notification_channels  JSONB NOT NULL DEFAULT '{"push":true,"email":true}',
    deleted_at             TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email ON auth.users (email);
CREATE UNIQUE INDEX idx_users_phone ON auth.users (phone) WHERE phone IS NOT NULL;

CREATE TABLE auth.oauth_identities (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    provider          TEXT NOT NULL CHECK (provider IN ('google')),
    provider_user_id  TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_oauth_identities_provider_provider_user_id
    ON auth.oauth_identities (provider, provider_user_id);

CREATE TABLE auth.verification_codes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    channel         TEXT NOT NULL CHECK (channel IN ('email', 'phone')),
    purpose         TEXT NOT NULL CHECK (purpose IN ('register', 'password_reset')),
    code_hash       TEXT NOT NULL,
    attempts_count  INT NOT NULL DEFAULT 0,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_verification_codes_user_id_channel_purpose
    ON auth.verification_codes (user_id, channel, purpose);

CREATE TABLE auth.refresh_tokens (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL,
    family_id    UUID NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ,
    replaced_by  UUID REFERENCES auth.refresh_tokens(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_token_hash ON auth.refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_family_id ON auth.refresh_tokens (family_id);
CREATE INDEX idx_refresh_tokens_user_id_expires_at ON auth.refresh_tokens (user_id, expires_at);

CREATE TABLE auth.children (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    institution_id       UUID NOT NULL REFERENCES catalog.institutions(id) ON DELETE RESTRICT,
    age_group            TEXT NOT NULL CHECK (age_group IN ('kindergarten', 'preschool', 'primary', 'basic', 'secondary', 'university')),
    status               TEXT NOT NULL CHECK (status IN ('current', 'alumnus', 'transferred')),
    confirmation_status  TEXT NOT NULL DEFAULT 'pending' CHECK (confirmation_status IN ('pending', 'confirmed', 'rejected')),
    confirmed_by         UUID,
    confirmed_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (user_id, institution_id)
);

CREATE INDEX idx_children_institution_id_confirmation_status
    ON auth.children (institution_id, confirmation_status);

-- +goose Down
DROP TABLE auth.children;
DROP TABLE auth.refresh_tokens;
DROP TABLE auth.verification_codes;
DROP TABLE auth.oauth_identities;
DROP TABLE auth.users;
